package fmt

type FmtInterface interface {
	Bold(s string) string
	Italic(b string) string
	Underline(b string) string
	Block(b string) string
	// LineSplitter A function that produces a string that carries one string to another.
	//
	// According to the standard </n>
	LineSplitter() string
	// InfoSplitter A function that produces a special line that carries one part of the message from another
	//
	// In my case, i use "============ * 1298121893" (WITHOUT LineSplitter())
	InfoSplitter() string
	Tab() string
	// Mention For mention a particular player.
	// <@nick>, for example.
	Mention(nick string) string
}

func BoldItalic(f FmtInterface, s string) string      { return f.Italic(f.Bold(s)) }
func BoldUnderline(f FmtInterface, s string) string   { return f.Underline(f.Bold(s)) }
func ItalicUnderline(f FmtInterface, s string) string { return f.Italic(f.Underline(s)) }

type NilFMTInterface struct{}

func (f NilFMTInterface) Bold(s string) string       { return s }
func (f NilFMTInterface) Italic(s string) string     { return s }
func (f NilFMTInterface) Underline(s string) string  { return s }
func (f NilFMTInterface) Block(s string) string      { return s }
func (f NilFMTInterface) LineSplitter() string       { return "\n" }
func (f NilFMTInterface) InfoSplitter() string       { return "\n\n" }
func (f NilFMTInterface) Tab() string                { return "\t" }
func (f NilFMTInterface) Mention(nick string) string { return nick }

var NilFMTInterfaceInstance = NilFMTInterface{}
