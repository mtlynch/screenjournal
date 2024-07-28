package screenjournal

type EmailBodyMarkdown string

func (m EmailBodyMarkdown) String() string {
	return string(m)
}
