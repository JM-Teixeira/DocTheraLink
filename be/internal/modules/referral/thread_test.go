package referral

import (
	"context"
	"testing"
)

// The conversation: a question stays open (badge + filter for whoever must
// answer) until the other side replies to it; progress notes carry pain.
func TestQuestionsAndProgress(t *testing.T) {
	ctx := context.Background()
	s, _ := newTestService(t)
	r := mustCreate(t, s, ptrUint(physio.ID))
	if _, err := s.Accept(ctx, physio, r.ID); err != nil {
		t.Fatal(err)
	}

	q, err := s.AddNote(ctx, doctor, r.ID, NoteRequest{Kind: NoteQuestion, Body: "Há derrame articular?"})
	if err != nil {
		t.Fatal(err)
	}
	open := func(a Actor) int {
		list, _, err := s.List(ctx, a, ListFilter{})
		if err != nil {
			t.Fatal(err)
		}
		for _, x := range list {
			if x.ID == r.ID {
				return x.OpenQuestions
			}
		}
		return -1
	}
	if open(physio) != 1 || open(doctor) != 0 {
		t.Fatalf("waiting for the physio only: physio %d, doctor %d", open(physio), open(doctor))
	}
	if list, _, _ := s.List(ctx, physio, ListFilter{OpenQuestions: true}); len(list) != 1 {
		t.Errorf("filter for the physio: %d", len(list))
	}
	if list, _, _ := s.List(ctx, doctor, ListFilter{OpenQuestions: true}); len(list) != 0 {
		t.Errorf("filter for the doctor: %d", len(list))
	}

	// Replies must point at a question asked by the other side
	if _, err := s.AddNote(ctx, physio, r.ID, NoteRequest{Kind: NoteReply, Body: "Sim"}); err == nil {
		t.Error("a reply without its question")
	}
	if _, err := s.AddNote(ctx, doctor, r.ID, NoteRequest{Kind: NoteReply, Body: "Eu próprio", ReplyTo: q.ID}); err == nil {
		t.Error("answering your own question")
	}
	pain := 7
	p, err := s.AddNote(ctx, physio, r.ID, NoteRequest{Kind: NoteProgress, Body: "Dor ao subir escadas", Pain: &pain})
	if err != nil || p.Pain == nil || *p.Pain != 7 {
		t.Fatalf("progress with pain: %+v %v", p, err)
	}
	if _, err := s.AddNote(ctx, physio, r.ID, NoteRequest{Kind: NoteReply, Body: "Sim", ReplyTo: p.ID}); err == nil {
		t.Error("replying to something that is not a question")
	}
	reply, err := s.AddNote(ctx, physio, r.ID, NoteRequest{Kind: NoteReply, Body: "Ligeiro, a diminuir", ReplyTo: q.ID})
	if err != nil || reply.ReplyToID == nil || *reply.ReplyToID != q.ID {
		t.Fatalf("reply: %+v %v", reply, err)
	}
	if open(physio) != 0 {
		t.Errorf("answered: still %d open", open(physio))
	}
	notes, _ := s.Notes(ctx, doctor, r.ID)
	for _, n := range notes {
		if n.ID == q.ID && !n.Answered {
			t.Error("the question shows as answered")
		}
	}
	// Pain belongs to progress notes only
	if n, _ := s.AddNote(ctx, doctor, r.ID, NoteRequest{Kind: NoteQuestion, Body: "E a dor?", Pain: &pain}); n.Pain != nil {
		t.Error("pain on a question")
	}
}
