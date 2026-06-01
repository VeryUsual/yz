package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/zenizh/go-capturer"
	"strings"
)

func TestMain(t *testing.T) {
	var verbose = false
	out := capturer.CaptureStdout(func () {
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

		`, &verbose)
	})
	assert.Equal(t, strings.TrimSpace(out), "203\nHello, world!\nx equals x\n1\n2")
}