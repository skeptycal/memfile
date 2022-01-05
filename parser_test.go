package memfile

import "testing"

func Test_process(t *testing.T) {
	testString := "Now is the time for ALL good men to come to the aid of their country."
	type args struct {
		c  Cases
		in string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"none", args{}, testString},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := process(tt.args.c, tt.args.in); got != tt.want {
				t.Errorf("process() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_caser_process(t *testing.T) {
	type fields struct {
		in  string
		out string
		c   Cases
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &caser{
				in:  tt.fields.in,
				out: tt.fields.out,
				c:   tt.fields.c,
			}
			ca.process()
		})
	}
}
