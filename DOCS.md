# Docs

## Introduction

YZ is an interpreted programming language designed to be fast, easy to use and learn, and simple.

Its design has been influenced by Go, Rust, C, JavaScript, and Python.

YZ is simple yet extremely powerful.

### Installation

Learn how to install YZ [here](README.md).

### Getting started

To start writing YZ code, just make a new file, call it anything, and use the file extension ".yz". Then, use the YZ package that you just installed to run it.

### Hello world

```
println("Hello, world!");
```

Save this to a file named `hello.yz`. Now run `yz run hello.yz`.

If it prints hello world, this means you have written your first YZ program! Congratulations!

From the example above, you can see that we are using a built-in function called `println` and providing a parameter of type string to it.

## Syntax

### Comments

```
// This is a comment.
// This is the only way to write comments.
```

### Functions

```
func add(first_number, second_number) {
    let result = first_number + second_number;
    return result;
}

func sub(first_number, #arbitrary_params_allowed) {
    return first_number - second_number;
}

let x = add(first_number 4, second_number 3);
let y = sub(first_number 5, second_number 8);
println(x + y);
```

To specify parameters in function calls, you must put the parameter name, space, then your desired value. The indicator `#arbitrary_params_allowed` inside a function declaration's parameters means that any parameters will be accepted, but, keep in mind, if those parameters are referenced to and not passed, there will be an error.

Also, you may be wondering, why does the `println` function not have a parameter name that is needed? This is because `println` is not a function, it is instead a built-in callable. Built-in callables look exactly like functions, but are directly built into the interpreter and therefore don't need parameters to be specified. Don't be worried though, because most functions come from the standard library, except `println` and `_yz_invoke`.

To return a value from a function, do `return value;`. `return;` is invalid code and will result in a syntax error. You must return always something. Returning nothing impedes expressiveness.

### Standard library

```
import stdlib;

x = 5;

println(x);
println("as a string is:");
println(str(val must_num(num x)));

println("Here's 5 random numbers between 1 and 100:");
println(rand_num(min 1, max 100));
println(rand_num(min 1, max 100));
println(rand_num(min 1, max 100));
println(rand_num(min 1, max 100));
println(rand_num(min 1, max 100));

println("And here's a big one");
println(rand_num(min 10000000, max 100000000));
```

The standard library contains the core functions needed in most YZ programs. You can see the standard library at [libs/stdlib.yz](libs/stdlib.yz).

Let's take a sample function from it:

```
func rand_num(min, max) public {
	let min = must_num(num min);
	let max = must_num(num max);
	_yz_invoke(_yz_cmd_rand_num, rand_num, min min, max max);
	return rand_num;
}
```

But wait, hold on, what even is this doing? What is this `_yz_invoke` function?

The YZ invoke function is how the standard library interacts with the Go interpreter. This also allows for multiple standard library interpretations, and allows the standard library to be small and compact. This also allows the language to do things that it can't, by calling Go. And yes, you can write all the `_yz_invoke`'s yourself if you want. It'll just be inefficient.

### Variables

```
let x = 0;
while x < 5 {
    println(str(val x));
    let x = x + 1;
}
```

Variables are declared using `let name = value`. Variables are also modified using `let name = value`, for the sake of simplicity and expressiveness. All variables are mutable.

### Types
There are 3 types in YZ. String, integer, dictionary, and array. String and integer are built-in. Dictionary and array rely on the standard library. You'll see this widespread reliance on the standard library a lot, and I think it helps the simplicity, speed, and modularity of the language. Imagine just being able to turn off dictionaries for teaching, learning, much more speed, or just needing a simple language? That's the power of YZ.

```
let mystr = "string";
let myint = 5;

import stdlib;

let mylist = List_Make();
List_Append(list mylist, value "hello");
List_Append(list mylist, value 1);
println(List_ValueFromIndex(list mylist, index 0));

let dictionary = Dict_Make();
Dict_Set(dict dictionary, key "apple", value "red");
Dict_Set(dict dictionary, key "banana", value "yellow");
println(Dict_Get(dict dictionary, key "apple"));
```

### A short stop
You might be thinking, wow, this language is so inefficient, wow, it's so much lines to write even the most simple things, wow, it's such a hassle. But it's expressive, it allows your code, to express in a way, that anyone can read, understand, and use. It's fast, it allows you to exclude language features you don't need. If you want to write things in the shortest possible way, make spaghetti code unreadable even to yourself, go look for another language. But, if you want modularity, speed, simplicity, and expressibility, choose YZ.

### If statements

```
if 5 == 5 {
    println("5 = 5");
}
if 3 < 2 {
    println("3 < 2");
}

if 2 == 2 {} else {
    println("this shouldn't happen!")
}
```

Looking at the third statement, you notice the lack of a `not` operator. This is intentional. `Not` operators are unneeded, because you should be accounting for all possibilities, therefore having something in the `if` statement and the `else` one too. If you truly don't, then you can simply use the `if condition {} else {code}` statement.

### Match statements

There are no match statements, as they are a shortcut to lessen the length of your code, and shortcuts are unneeded in YZ.

### While loops

```
let x = 1;
while x < 4 {
    println(str(val x));
    let x = x + 1;
}
```

### Gothru loops

Gothru loops gothru every element in an array.

```
let arr = List_Make();
List_Append(list mylist, value "hello");
List_Append(list mylist, value 1);
List_Append(list mylist, value "saluton");
List_Append(list mylist, value 2);
List_Append(list mylist, value "bonjour");
List_Append(list mylist, value 3);

gothru arr as element {
    println(element);
}
```

### Break statements

```
let x = 1;
while x < 4 {
    println(str(val x));
    let x = x + 1;
    if x == 2 {
        break;
    }
}
```

Breaks terminate the current loop.

### Imports

```
import stdlib;
import fs;
import guitk;
import requests;
```

Imports import libraries from the libs/ directory. After importing, you can use their public functions and variables.

### Math

```
println(4 + 3 * 2);
```

Math is possible anywhere numbers are allowed.

## Ending

You are now proficent in the YZ programming language. Refer back here as needed.