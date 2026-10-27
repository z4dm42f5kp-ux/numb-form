package app

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestTree(t *testing.T) {
	q := &request{To: "F", Others: 3, Changes: []change{
		{kind: kReplace, rel: "3. Übersichten/Liste.xlsx"},
		{kind: kAdd, rel: "Bar Ltda/Appendix A - Bar Ltda.docx", top: "Bar Ltda", origin: byNumbForm},
		{kind: kDup, rel: "Foo S.A/Appendix A - Foo S.A.docx", top: "Foo S.A", dup: "D10629 Foo S.A."},
		{kind: kKeep, rel: "numb-form.log"},
	}}
	got := q.tree()
	before := "Before: F\n├── 3. Übersichten\n│   └── Liste.xlsx\n├── D10629 Foo S.A.\n└── … 3 more, unchanged\n"
	if !strings.HasPrefix(got, before) {
		t.Fatalf("before:\n%s", got)
	}
	after := strings.SplitN(got, "After the commit:\n", 2)[1]
	lines := strings.Split(strings.TrimSpace(after), "\n")
	want := []string{"├── 3. Übersichten", "│   └── Liste.xlsx   ", "├── Bar Ltda   ", "│   └── Appendix A - Bar Ltda.docx   + new, made by numb-form, unchanged",
		"├── D10629 Foo S.A.   ", "└── … 3 more, unchanged"}
	if len(lines) != len(want) {
		t.Fatalf("after:\n%s", after)
	}
	for i, w := range want {
		if !strings.HasPrefix(lines[i], w) {
			t.Errorf("line %d = %q, want it to start with %q", i, lines[i], w)
		}
	}
	if strings.Contains(got, "numb-form.log") {
		t.Error("residue in the tree")
	}
}

// A request written to disk reads back as the same request, and its text
// part says what changes.
func TestRequestRoundTrip(t *testing.T) {
	real, cp := scanCopy(t)
	ctx := context.Background()
	rc, err := reviewSetup(ctx, cp, real, testConfig(t, filepath.Dir(cp)))
	if err != nil {
		t.Fatal(err)
	}
	q, err := scanRequest(ctx, cp, real, rc)
	if err != nil {
		t.Fatal(err)
	}
	q.Made = time.Date(2026, 10, 16, 9, 30, 0, 0, time.UTC)
	p := filepath.Join(cp, RequestName)
	if err := q.write(p); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Status: waiting", "New folder  Foo Ltd", "made by numb-form, unchanged",
		"a form edited after filling, made by hand or filled on another computer", "Replaced  ", "Left alone", "After the commit:", "+ new folder", "~ replaced"} {
		if !strings.Contains(string(b), want) {
			t.Errorf("request lacks %q:\n%s", want, b)
		}
	}
	got, err := readRequest(p)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Made.Equal(q.Made) {
		t.Errorf("made %v, want %v", got.Made, q.Made)
	}
	got.Made = q.Made
	if !reflect.DeepEqual(got, q) {
		t.Errorf("read back\n%+v\nwant\n%+v", got, q)
	}
	if _, err := os.Stat(p + ".tmp"); !os.IsNotExist(err) {
		t.Error("temporary file left behind")
	}
}

func TestReadRequestRefusesUnsafePaths(t *testing.T) {
	head := "text\n" + requestData + "\nfrom\tA\nto\tB\n"
	for _, line := range []string{
		"add\t../x.docx\t\t\t1\t\t\t\t",
		"add\t/x.docx\t\t\t1\t\t\t\t",
		"add\ta/../../x.docx\t\t\t1\t\t\t\t",
		"add\tFoo/x.docx\t\t\t1\tBar\t\t\t",
		"dup\tFoo/x.docx\t\t\t1\tFoo\t../Bar\t\t",
		"delete\tFoo/x.docx\t\t\t1\t\t\t\t",
	} {
		if _, err := parseRequest("r", []byte(head+line+"\n")); err == nil {
			t.Errorf("accepted %q", line)
		}
	}
	if _, err := parseRequest("r", []byte(head+"add\tFoo/x.docx\t\t\t1\tFoo\t\t\t\n")); err != nil {
		t.Error(err)
	}
	if _, err := parseRequest("r", []byte("text only\n")); err == nil {
		t.Error("accepted a request without data")
	}
}

func FuzzReadRequest(f *testing.F) {
	q := &request{From: "A", To: "B", Status: "waiting", Sheet: "3. Übersichten/Liste.xlsx", Changes: []change{
		{kind: kAdd, rel: "Foo Ltd/Appendix A - Foo Ltd.docx", sha: "ab", size: 10, top: "Foo Ltd", origin: byNumbForm},
		{kind: kReplace, rel: "3. Übersichten/Liste.xlsx", sha: "cd", old: "ef", size: 20},
		{kind: kKeep, rel: "numb-form.log", reason: "not copied: a log or system file"},
	}}
	p := filepath.Join(f.TempDir(), RequestName)
	if err := q.write(p); err != nil {
		f.Fatal(err)
	}
	seed, err := os.ReadFile(p)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte(requestData + "\nfrom\tA\nto\tB\nadd\t../x\t\t\t1\t\t\t\t\n"))
	f.Fuzz(func(t *testing.T, b []byte) {
		q, err := parseRequest("fuzz", b)
		if err != nil {
			return
		}
		if q.From == "" || q.To == "" {
			t.Fatal("a request without folders was accepted")
		}
		for _, c := range q.Changes {
			if !safeRel(c.rel) || (c.dup != "" && !safeRel(c.dup)) || (c.top != "" && !strings.HasPrefix(c.rel, c.top+"/")) {
				t.Fatalf("unsafe change accepted: %+v", c)
			}
		}
	})
}
