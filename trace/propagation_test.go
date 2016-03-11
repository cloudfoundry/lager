package trace_test

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pivotal-golang/lager/trace"
)

type verbatimCarrier struct {
	trace.Context
	b map[string]string
}

var _ trace.DelegatingCarrier = &verbatimCarrier{}

func (vc *verbatimCarrier) SetBaggageItem(k, v string) {
	vc.b[k] = v
}

func (vc *verbatimCarrier) GetBaggage(f func(string, string)) {
	for k, v := range vc.b {
		f(k, v)
	}
}

func (vc *verbatimCarrier) SetState(tID, sID int64, sampled bool) {
	vc.Context = trace.Context{TraceID: tID, SpanID: sID, Sampled: sampled}
}

func (vc *verbatimCarrier) State() (traceID, spanID int64, sampled bool) {
	return vc.Context.TraceID, vc.Context.SpanID, vc.Context.Sampled
}

func TestSpanPropagator(t *testing.T) {
	const op = "test"
	recorder := trace.NewInMemoryRecorder()
	tracer := trace.New(recorder)

	sp := tracer.StartSpan(op)
	sp.SetBaggageItem("foo", "bar")

	tests := []struct {
		typ, carrier interface{}
	}{
		{trace.Delegator, trace.DelegatingCarrier(&verbatimCarrier{b: map[string]string{}})},
		{opentracing.SplitBinary, opentracing.NewSplitBinaryCarrier()},
		{opentracing.SplitText, opentracing.NewSplitTextCarrier()},
		{opentracing.GoHTTPHeader, http.Header{}},
	}

	for i, test := range tests {
		if err := tracer.Inject(sp, test.typ, test.carrier); err != nil {
			t.Fatalf("%d: %v", i, err)
		}
		child, err := tracer.Join(op, test.typ, test.carrier)
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}
		child.Finish()
	}
	sp.Finish()

	spans := recorder.GetSpans()
	if a, e := len(spans), len(tests)+1; a != e {
		t.Fatalf("expected %d spans, got %d", e, a)
	}

	// The last span is the original one.
	exp, spans := spans[len(spans)-1], spans[:len(spans)-1]
	exp.Duration = time.Duration(123)
	exp.Start = time.Time{}.Add(1)

	for i, sp := range spans {
		if a, e := sp.ParentSpanID, exp.SpanID; a != e {
			t.Fatalf("%d: ParentSpanID %d does not match expectation %d", i, a, e)
		} else {
			// Prepare for comparison.
			sp.SpanID, sp.ParentSpanID = exp.SpanID, 0
			sp.Duration, sp.Start = exp.Duration, exp.Start
		}
		if a, e := sp.TraceID, exp.TraceID; a != e {
			t.Fatalf("%d: TraceID changed from %d to %d", i, e, a)
		}
		if !reflect.DeepEqual(exp, sp) {
			t.Fatalf("%d: wanted %+v, got %+v", i, spew.Sdump(exp), spew.Sdump(sp))
		}
	}
}
