package dashboard

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/opsybot/opsybot/internal/entity"
)

func multipartBody(t *testing.T, write func(*multipart.Writer)) *multipart.Reader {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	write(w)
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return multipart.NewReader(&buf, w.Boundary())
}

func TestReadAttachmentUploadAcceptsOneFile(t *testing.T) {
	reader := multipartBody(t, func(w *multipart.Writer) {
		_ = w.WriteField("label", "queue depth")
		part, err := w.CreateFormFile("file", "graph.png")
		if err != nil {
			t.Fatalf("create file part: %v", err)
		}
		_, _ = part.Write([]byte("not-really-a-png"))
	})

	upload, body, err := readAttachmentUpload(reader)
	if err != nil {
		t.Fatalf("read upload: %v", err)
	}
	if upload.Label != "queue depth" {
		t.Errorf("label = %q", upload.Label)
	}
	if upload.SizeBytes != int64(len("not-really-a-png")) {
		t.Errorf("size = %d", upload.SizeBytes)
	}
	got, _ := io.ReadAll(body)
	if string(got) != "not-really-a-png" {
		t.Errorf("body = %q", got)
	}
}

func TestReadAttachmentUploadRejectsRepeatedFileParts(t *testing.T) {
	const partSize = 1 << 20
	chunk := bytes.Repeat([]byte("x"), partSize)

	reader := multipartBody(t, func(w *multipart.Writer) {
		for i := 0; i < 5; i++ {
			part, err := w.CreateFormFile("file", "chunk.bin")
			if err != nil {
				t.Fatalf("create file part %d: %v", i, err)
			}
			_, _ = part.Write(chunk)
		}
	})

	_, _, err := readAttachmentUpload(reader)
	if !errors.Is(err, entity.ErrAttachmentUploadInvalid) {
		t.Fatalf("err = %v, want ErrAttachmentUploadInvalid; repeated file parts accumulate in memory", err)
	}
}

func TestReadAttachmentUploadRejectsOversizeFile(t *testing.T) {
	reader := multipartBody(t, func(w *multipart.Writer) {
		part, err := w.CreateFormFile("file", "huge.bin")
		if err != nil {
			t.Fatalf("create file part: %v", err)
		}
		_, _ = io.Copy(part, io.LimitReader(neverEnding{}, entity.AttachmentUploadMaxBytes+1024))
	})

	_, _, err := readAttachmentUpload(reader)
	if !errors.Is(err, entity.ErrAttachmentTooLarge) {
		t.Fatalf("err = %v, want ErrAttachmentTooLarge", err)
	}
}

func TestReadAttachmentUploadRejectsMissingFile(t *testing.T) {
	reader := multipartBody(t, func(w *multipart.Writer) {
		_ = w.WriteField("label", "no file here")
	})

	_, _, err := readAttachmentUpload(reader)
	if !errors.Is(err, entity.ErrAttachmentUploadInvalid) {
		t.Fatalf("err = %v, want ErrAttachmentUploadInvalid", err)
	}
}

func TestReadAttachmentUploadRejectsUnknownPart(t *testing.T) {
	reader := multipartBody(t, func(w *multipart.Writer) {
		part, err := w.CreateFormFile("file", "graph.png")
		if err != nil {
			t.Fatalf("create file part: %v", err)
		}
		_, _ = part.Write([]byte("ok"))
		_ = w.WriteField("objectKey", "incidents/other-workspace/evil")
	})

	_, _, err := readAttachmentUpload(reader)
	if !errors.Is(err, entity.ErrAttachmentUploadInvalid) {
		t.Fatalf("err = %v, want ErrAttachmentUploadInvalid", err)
	}
}

func TestReadAttachmentUploadFallsBackToFileName(t *testing.T) {
	reader := multipartBody(t, func(w *multipart.Writer) {
		part, err := w.CreateFormFile("file", "dashboard.png")
		if err != nil {
			t.Fatalf("create file part: %v", err)
		}
		_, _ = part.Write([]byte("bytes"))
	})

	upload, _, err := readAttachmentUpload(reader)
	if err != nil {
		t.Fatalf("read upload: %v", err)
	}
	if upload.Label != "dashboard.png" {
		t.Errorf("label = %q, want the file name", upload.Label)
	}
}

func TestReadAttachmentUploadTruncatesOverlongLabel(t *testing.T) {
	long := strings.Repeat("l", entity.AttachmentLabelMaxLength*2)
	reader := multipartBody(t, func(w *multipart.Writer) {
		_ = w.WriteField("label", long)
		part, err := w.CreateFormFile("file", "graph.png")
		if err != nil {
			t.Fatalf("create file part: %v", err)
		}
		_, _ = part.Write([]byte("ok"))
	})

	upload, _, err := readAttachmentUpload(reader)
	if err != nil {
		t.Fatalf("read upload: %v", err)
	}
	if len(upload.Label) > entity.AttachmentLabelMaxLength+1 {
		t.Fatalf("label buffered %d bytes unbounded", len(upload.Label))
	}
	if err := (entity.NewAttachment{
		Kind: entity.AttachmentImage, Label: upload.Label, ContentType: "image/png",
	}).Validate(); err == nil {
		t.Fatal("an overlong label must fail entity validation rather than be silently accepted")
	}
}

type neverEnding struct{}

func (neverEnding) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'x'
	}
	return len(p), nil
}
