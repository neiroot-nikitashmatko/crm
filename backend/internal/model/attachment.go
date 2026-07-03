package model

const (
	AttachmentEntityDeal = "deal"
	AttachmentEntityTask = "task"
	AttachmentEntityLead = "lead"
)

const MaxAttachmentSizeBytes = 10 * 1024 * 1024

type Attachment struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	MimeType   string `json:"mimeType"`
	UploadedBy string `json:"uploadedBy"`
	UploadedAt int64  `json:"uploadedAt"`
}

type AttachmentContent struct {
	Attachment
	Content []byte
}
