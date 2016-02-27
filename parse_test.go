// Copyright ©2016 Jonathan J Lawlor. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parsefloat

import (
	"errors"
	"math"
	"regexp"
	"testing"
)

func TestNew(t *testing.T) {
	for i, tt := range []struct {
		varre string
		expr  string
		rpn   []string
		str   string
		vars  map[string]float64
		want  float64
	}{
		{ // constant
			varre: `(?P<N>\d+)-\d+$`,
			expr:  "1.0",
			rpn:   []string{"1.0"},
			str:   "1.0",
			vars:  nil,
			want:  1.0,
		},
		{ // variable
			varre: `(?P<N>\d+)-\d+$`,
			expr:  "N",
			rpn:   []string{"N"},
			str:   "N",
			vars:  map[string]float64{"N": 10.0},
			want:  10.0,
		},
		{ // multiplication
			varre: `(?P<N>\d+)-\d+$`,
			expr:  "N*N",
			rpn:   []string{"N", "N", "*"},
			str:   "N*N",
			vars:  map[string]float64{"N": 10.0},
			want:  100.0,
		},
		{ // unary function of N
			varre: `(?P<N>\d+)-\d+$`,
			expr:  "math.Log(N)",
			rpn:   []string{"N", "math.Log"},
			str:   "math.Log(N)",
			vars:  map[string]float64{"N": 10.0},
			want:  math.Log(10.0),
		},
		{ // binary function of two different inputs
			varre: `(?P<M>\d+)(?P<N>\d+)-\d+$`,
			expr:  "-math.Hypot(M+N, M-N)",
			rpn:   []string{"M", "N", "+", "M", "N", "-", "math.Hypot", "u-"},
			str:   "-math.Hypot(M+N, M-N)",
			vars:  map[string]float64{"M": 3.5, "N": 0.5},
			want:  -5.0,
		},
		{ // unary plus and division
			varre: `(?P<M>\d+)(?P<N>\d+)-\d+$`,
			expr:  "+M/N",
			rpn:   []string{"M", "u+", "N", "/"},
			str:   "+M/N",
			vars:  map[string]float64{"M": 3.5, "N": 0.5},
			want:  7.0,
		},
	} {
		re := regexp.MustCompile(tt.varre)
		names := NamedVars(re)
		expr, err := New(tt.expr, names)
		if err != nil {
			panic(err)
		}
		if x := expr.Eval(tt.vars); x != tt.want {
			t.Errorf("%d: expected %s = %f, got %f", i, expr, tt.want, x)
		}
		if x := expr.String(); x != tt.str {
			t.Errorf("%d: expected %g.String() = %s, got %s", i, expr, tt.str, x)
		}
		for xouti, xout := range expr.(*expression).output {
			if xout.String() != tt.rpn[xouti] {
				t.Errorf("%d: expected (%s).output[%d].String() = %s, got %s", i, expr, xouti, tt.rpn[xouti], xout.String())
			}
		}
	}
}

func TestNewSlice(t *testing.T) {
	for i, tt := range []struct {
		varre string
		expr  string
		vars  map[string]float64
		want  []float64
	}{
		{ // constant
			varre: `(?P<N>\d+)-\d+$`,
			expr:  "float64{1.0}",
			vars:  nil,
			want:  []float64{1.0},
		},
		{ // quadratic
			varre: `(?P<N>\d+)-\d+$`,
			expr:  "float64{N*N, N, 1.0}",
			vars:  map[string]float64{"N": 10.0},
			want:  []float64{100.0, 10.0, 1.0},
		},
		{ // n log n
			varre: `(?P<N>\d+)-\d+$`,
			expr:  "float64{N*math.Log(N), 1.0}",
			vars:  map[string]float64{"N": 10.0},
			want:  []float64{math.Log(10.0) * 10.0, 1.0},
		},
	} {
		re := regexp.MustCompile(tt.varre)
		names := NamedVars(re)
		expr, err := NewSlice(tt.expr, names)
		if err != nil {
			panic(err)
		}
		for xi, ex := range expr {
			if x := ex.Eval(tt.vars); x != tt.want[xi] {
				t.Errorf("%d: expected %s[%d] = %f, got %f", i, ex, xi, tt.want[xi], x)
			}
		}
	}
}

func TestNewErr(t *testing.T) {
	for i, tt := range []struct {
		expr    string
		names   map[string]struct{}
		wanterr error
	}{
		{
			expr:    "(()", // invalid nesting
			names:   map[string]struct{}{},
			wanterr: errors.New("1:3: expected operand, found ')'"),
		},
		{
			expr:    "N + 1.0", // unknown variable
			names:   map[string]struct{}{},
			wanterr: errors.New("unknown variable: N"),
		},
	} {
		_, err := New(tt.expr, tt.names)
		if err.Error() != tt.wanterr.Error() {
			if err == nil {
				t.Errorf("%d: expected err=%s, got nil", i, tt.wanterr)
				continue
			}
			t.Errorf("%d: expected err=%s, got %s", i, tt.wanterr, err)
		}
	}
}

func TestNewSliceErr(t *testing.T) {
	for i, tt := range []struct {
		expr    string
		names   map[string]struct{}
		wanterr error
	}{
		{
			expr:    "(()", // invalid nesting
			names:   map[string]struct{}{},
			wanterr: errors.New("1:3: expected operand, found ')'"),
		},
		{
			expr:    "N + 1.0", // unknown variable
			names:   map[string]struct{}{},
			wanterr: errors.New("expression N + 1.0 is not a []float64"),
		},
	} {
		_, err := NewSlice(tt.expr, tt.names)
		if err.Error() != tt.wanterr.Error() {
			if err == nil {
				t.Errorf("%d: expected err=%s, got nil", i, tt.wanterr)
				continue
			}
			t.Errorf("%d: expected err=%s, got %s", i, tt.wanterr, err)
		}
	}
}
