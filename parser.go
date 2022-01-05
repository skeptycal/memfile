package memfile

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

var nlbyte []byte = []byte{'\n'}

func NewParser(r io.Reader) (Parser, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	p := parser{}
	p.input = string(b) // initialize other fields JIT as needed

	return &parser{}, nil
}

// func NewParser(p []byte) Parser {
// 	return &parser{input: p}
// }

type Parser interface {
	Parse() (string, error)
}

// ParserOptions contains the configuration information
// for the Parser.
//
// These fields are only used if the behavior is desired,
// otherwise they should not be set.
//
// If lineseps are provided, they will be converted to the
// default linesep (or nl if it is given)
type ParserOptions struct {
	nl            string            // single newline character to use (default \n)
	sep           string            // single field delimiter to use (default \t)
	characterCase Cases             // none, upper, lower, title, camel, snake, Pascal, kehab
	lineseps      []string          // list of acceptable line delimiters (default empty)
	fieldseps     []string          // list of acceptable field delimiters (default empty)
	prefix        string            // prefix to remove (default "")
	addprefix     string            // prefix to add (default "")
	suffix        string            // suffix to remove (default "")
	addsuffix     string            // suffix to add (default "")
	cut           []string          // strings to remove (default empty)
	replace       map[string]string // map of strings to replace (default empty)
}

type Cases byte

const (
	none         Cases = iota // Now is the time for ALL good men to come to the aid of their country.
	upper                     // NOW IS THE TIME FOR ALL GOOD MEN TO COME TO THE AID OF THEIR COUNTRY.
	lower                     // now is the time for all good men to come to the aid of their country.
	title                     // Now Is The Time For All Good Men To Come To The Aid Of Their Country.
	reverse                   // nOW IS THE TIME FOR all GOOD MEN TO COME TO THE AID OF THEIR COUNTRY.
	camel                     // nowIsTheTimeForALLGoodMenToComeToTheAidOfTheirCountry.
	snake                     // now_is_the_time_for_all_good_men_to_come_to_the_aid_of_their_country.
	snakeAllCaps              // NOW_IS_THE_TIME_FOR_ALL_GOOD_MEN_TO_COME_TO_THE_AID_OF_THEIR_COUNTRY.
	Pascal                    // NowIsTheTimeForALLGoodMenToComeToTheAidOfTheirCountry.
	kehab                     // now-is-the-time-for-all-good-men-to-come-to-the-aid-of-their-country.
	snakeCamel                // now_Is_The_Time_For_All_Good_Men_To_Come_To_The_Aid_Of_Their_Country.
	snakePascal               // Now_Is_The_Time_For_All_Good_Men_To_Come_To_The_Aid_Of_Their_Country.
	kehabCamel                // now-Is-The-Time-For-All-Good-Men-To-Come-To-The-Aid-Of-Their-Country
	kehabPascal               // Now-Is-The-Time-For-All-Good-Men-To-Come-To-The-Aid-Of-Their-Country
)

type Caser interface {
	String() string
	ChangeCase(c Cases)
}

type caser struct {
	in  string
	out string
	c   Cases
}

// String returns a human-readable representation
// of the string with the corresponding case.
func (ca *caser) String() string {
	if ca.out == "" {
		ca.process()
	}
	return ca.out
}

// ChangeCase changes the output case of the string.
func (ca *caser) ChangeCase(c Cases) {
	ca.c = c
	ca.process()
}

func fold(in string) string {
	sb := strings.Builder{}
	defer sb.Reset()

	for _, r := range in {
		sb.WriteRune(unicode.SimpleFold(r))
	}
	return sb.String()
}

func spacer(in, repl string, firstLower, titleCase bool) (retval string) {

	if titleCase {
		in = strings.ToTitle(in)
	}

	if firstLower {
		retval = strings.ToLower(in[:1])
	} else {
		retval = in[1:]
	}

	retval += strings.ReplaceAll(in[1:], " ", repl)

	return retval
}

func ToSnake(in string) string {
	return spacer(strings.ToLower(in), "_", true, false)
}

func ToPascal(in string) string {
	return ""
}

func process(c Cases, in string) string {
	switch c {
	case none:
		return ""
	case upper: // NOW IS THE TIME FOR ALL GOOD MEN TO COME TO THE AID OF THEIR COUNTRY.
		return strings.ToUpper(in)
	case lower: // now is the time for all good men to come to the aid of their country.
		return strings.ToLower(in)
	case title: // Now Is The Time For All Good Men To Come To The Aid Of Their Country.
		return strings.ToTitle(in)
	case reverse: // nOW IS THE TIME FOR all GOOD MEN TO COME TO THE AID OF THEIR COUNTRY.
		return fold(in)
	case camel: // nowIsTheTimeForALLGoodMenToComeToTheAidOfTheirCountry.
		return spacer(in, "", true, true)
	case snake: // now_is_the_time_for_all_good_men_to_come_to_the_aid_of_their_country.
		return spacer(strings.ToLower(in), "_", true, false)
	case snakeAllCaps: // NOW_IS_THE_TIME_FOR_ALL_GOOD_MEN_TO_COME_TO_THE_AID_OF_THEIR_COUNTRY.
		return spacer(strings.ToUpper(in), "_", true, false)
	case Pascal: // NowIsTheTimeForALLGoodMenToComeToTheAidOfTheirCountry.
		return spacer(in, "", false, true)
	case kehab: // now-is-the-time-for-all-good-men-to-come-to-the-aid-of-their-country.
		return spacer(strings.ToLower(in), "-", true, false)
	case snakeCamel: // now_Is_The_Time_For_All_Good_Men_To_Come_To_The_Aid_Of_Their_Country.
		return spacer(strings.ToLower(in), "_", true, true)
	case snakePascal: // Now_Is_The_Time_For_All_Good_Men_To_Come_To_The_Aid_Of_Their_Country.
		return spacer(strings.ToLower(in), "_", false, true)
	case kehabCamel: // now-Is-The-Time-For-All-Good-Men-To-Come-To-The-Aid-Of-Their-Country
		return spacer(strings.ToLower(in), "-", true, true)
	case kehabPascal: // Now-Is-The-Time-For-All-Good-Men-To-Come-To-The-Aid-Of-Their-Country
		return spacer(strings.ToLower(in), "-", false, true)
	default:
		return in
	}
}

func (ca *caser) process() {
	ca.out = process(ca.c, ca.in)
}

func NewCaseStringer(c Cases, s string) Caser {
	return &caser{s, "", c}
}

type ParserFunctionMap map[string]func()

type parser struct {
	input  string           // original input string
	lines  []string         // buffer for lines of text
	output *strings.Builder // output builder

	options ParserOptions

	dirty bool // is output modified
}

func (p *parser) Parse() (string, error) {
	return p.parse()
}

// parse contains the specific parsing algorithm
func (p *parser) parse() (string, error) {

	// var skipPrefixes []byte = []byte{'#'}

	// process 'whole text' changes
	p.setLineBreak("\n")

	// split into lines
	p.split()

	// process 'line by line' changes

	// for _, line := range bytes.Split(b, nlbyte) {
	// 	if bytes.HasPrefix(line, skipPrefixes) {
	// 		continue
	// 	}
	// }
	return p.output.String(), nil
}

func (p *parser) split() error {
	if len(p.lines) > 0 {
		return errors.New("parser.split may only be run once")
	}
	// p.lines = strings.Split(p.input, p.nl)
	return nil
}

func (p *parser) checklines() error {
	if len(p.lines) < 1 {
		err := p.split()
		if err != nil {
			return fmt.Errorf("parser could not access lines: %v", err)
		}
	}
	return nil
}

func (p *parser) setLineBreak(nl string) {
	p.options.nl = nl
}

func (p *parser) replaceLineDelimeters(old, new string) {
	strings.ReplaceAll(p.input, old, p.options.sep)
	return
}

func (p *parser) removeSuffix() error {
	return nil
}

func (p *parser) removePrefix() error {

	for i, line := range p.lines {
		p.lines[i] = strings.TrimPrefix(line, p.options.prefix)
	}
	return nil
}

func (p *parser) skipPrefixes() error {
	return nil
}
