package cli

import (
	"reflect"
	"testing"
)

func Test_history_append(t *testing.T) {
	type fields struct {
		values []string
		idx    int
		cap    int
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *history
	}{
		{
			name: "Add to empty buffer",
			fields: fields{
				values: []string{"", "", ""},
				idx:    2,
				cap:    3,
			},
			args: args{s: "a"},
			want: &history{values: []string{"a", "", ""}, idx: 0, cap: 3},
		},
		{
			name: "Add to half filled buffer",
			fields: fields{
				values: []string{"a", "b", ""},
				idx:    1,
				cap:    3,
			},
			args: args{s: "c"},
			want: &history{values: []string{"a", "b", "c"}, idx: 2, cap: 3},
		},
		{
			name: "Add to full buffer",
			fields: fields{
				values: []string{"a", "b", "c"},
				idx:    2,
				cap:    3,
			},
			args: args{s: "d"},
			want: &history{values: []string{"d", "b", "c"}, idx: 0, cap: 3},
		},
		{
			name: "Add to wrapped around buffer",
			fields: fields{
				values: []string{"d", "b", "c"},
				idx:    0,
				cap:    3,
			},
			args: args{s: "e"},
			want: &history{values: []string{"d", "e", "c"}, idx: 1, cap: 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &history{
				values: tt.fields.values,
				idx:    tt.fields.idx,
				cap:    tt.fields.cap,
			}
			h.append(tt.args.s)

			if !reflect.DeepEqual(h, tt.want) {
				t.Errorf("Incorrect history post append")
			}
		})
	}
}

func Test_history_get(t *testing.T) {
	type fields struct {
		values []string
		idx    int
		cap    int
	}
	type args struct {
		offset uint
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Get from empty history",
			fields: fields{
				values: []string{"", "", ""},
				idx:    0,
				cap:    3,
			},
			args: args{offset: 0},
			want: "",
		},
		{
			name: "Get current element",
			fields: fields{
				values: []string{"d", "b", "c"},
				idx:    0,
				cap:    3,
			},
			args: args{offset: 0},
			want: "d",
		},
		{
			name: "Wrap around to end",
			fields: fields{
				values: []string{"d", "b", "c"},
				idx:    0,
				cap:    3,
			},
			args: args{offset: 1},
			want: "c",
		},
		{
			name: "Can wrap multiple times",
			fields: fields{
				values: []string{"d", "b", "c"},
				idx:    0,
				cap:    3,
			},
			args: args{offset: 5},
			want: "b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &history{
				values: tt.fields.values,
				idx:    tt.fields.idx,
				cap:    tt.fields.cap,
			}
			if got := h.get(tt.args.offset); got != tt.want {
				t.Errorf("history.get() = %v, want %v", got, tt.want)
			}
		})
	}
}
