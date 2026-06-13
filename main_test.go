package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/zenizh/go-capturer"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	var verbose = false
	out := capturer.CaptureStdout(func() {
		run_program(`

		let x = 4;
		let y = 4 + 3 * 7 + x;
		println(y * 7);
		println("Hello, world!");

		let x = 5;
		let f = "qwertyuiopasdfghjklzxcvbnm";
		if x == f {
			println("x equals f");
		}
		if x == x {
			println("x equals x");
		}

		func two() {
			println("2");
		}

		func one() {
			println("1");
		}

		one();
		two();


		// add function - adds 2 numbers
		func add(first_number, second_number) {
			let result = first_number + second_number;
			return result;
		}

		let x = add(first_number 4, second_number 4);
		println("Four plus four is:");
		println(x);

		func sub(#arbitrary_params_allowed) {
			return first - second;
		}
		println(sub(first 4, second 4));

		let four = 4;
		if four + four == 9 {
			println("four plus four is nine!");
		} else {
			println("four plus four isn't nine");
		}

		`, &verbose)
	})
	assert.Equal(t, strings.TrimSpace(out), "203\nHello, world!\nx equals x\n1\n2\nFour plus four is:\n8\n0\nfour plus four isn't nine")
}
