package dialogue

import "testing"

func TestParseRegistryValidatesThen(t *testing.T) {
	good := `{"dialogues":{"a":{"speaker":"x","pages":["hi"],"then":"b"},"b":{"speaker":"x","pages":["again"]}}}`
	r, err := ParseRegistry([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	if d, ok := r.Get("a"); !ok || d.Then != "b" {
		t.Fatalf("got %+v", d)
	}

	bad := `{"dialogues":{"a":{"speaker":"x","pages":["hi"],"then":"zzz"}}}`
	if _, err := ParseRegistry([]byte(bad)); err == nil {
		t.Fatal("dangling then should be rejected")
	}
	empty := `{"dialogues":{"a":{"speaker":"x","pages":[]}}}`
	if _, err := ParseRegistry([]byte(empty)); err == nil {
		t.Fatal("no pages should be rejected")
	}
}

func TestRunnerConfirmSequence(t *testing.T) {
	r := NewRunner(300)
	if r.Active() {
		t.Fatal("should start idle")
	}
	d := &Def{Speaker: "s", Pages: []string{"one", "two"}, Give: []Gift{{"coin", 3}}}
	var gifts []Gift
	var ended *Def
	r.OnGive = func(g Gift) { gifts = append(gifts, g) }
	r.OnEnd = func(d *Def) { ended = d }

	r.Start(d)
	if !r.Active() || r.Box.Done() {
		t.Fatal("Start should show page 0 with the reveal running")
	}
	r.Confirm() // finishes the reveal, does not turn the page
	if !r.Box.Done() || r.page != 0 {
		t.Fatal("first confirm should only finish the page")
	}
	r.Confirm() // turns to page 1
	if r.page != 1 || !r.Active() {
		t.Fatal("second confirm should turn the page")
	}
	r.Box.Finish()
	r.Confirm() // past the last page: ends, gives
	if r.Active() {
		t.Fatal("dialogue should have ended")
	}
	if len(gifts) != 1 || gifts[0].Count != 3 || ended != d {
		t.Fatalf("gifts %v ended %v", gifts, ended)
	}
	r.Confirm() // idle: no-op, no panic
}
