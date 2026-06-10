# YZ Programming Language

YZ is a fast, easy-to-learn, robust programming language. Its compiler is written in Go.

## Features
- Fast execution (it is ~28% faster than Python, ~35% faster than Ruby, ~36% faster than Node.JS, ~56% faster than PHP)
- Intuitive syntax
- Variables
- Functions
- Types
- Conditional statements and loops
- Error handling
- Standard library
- Filesystem library
- GUI library
- HTTP request library
- Cross-platform compatable

## Installation
```bash
git clone https://github.com/VeryUsual/yz
go build main.go
./main examples/0.yz
```

## Examples
### Hello world
```
println("Hello, world!");
```
Find more examples in the `examples/` directory.

## Syntax
### Variables
```
let x = 5;
let y = 6;
let z = x + y;
println(z);
let z = 4;
```
Variables are modified using `let var = value`, `var = value` is incorrect syntax.
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
```
To specify parameters in function calls, you must put the parameter name, space, then your desired value. The indicator `#arbitrary_params_allowed` inside a function
declaration's parameters means that any parameters will be accepted, but, keep in mind, if those parameters are referenced to and not passed, there will be an error.
### Conditions
```
if 5 >= 3 {
    println("5 is larger or equal to 3");
}
```
### Loops
```
let x = 4;
while x < 5 {
    println(x);
    let x = x + 1;
}


import stdlib;

let arr = List_Make();

List_Append(list arr, value "one");
List_Append(list arr, value 2);
List_Append(list arr, value "three");

gothru arr as element {
    println(element);
}
```

## License
This project is licensed under the GNU General Public License 3.0.