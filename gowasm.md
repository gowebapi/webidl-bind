# Go Web Assembly target

The output should be considers as helping libraries to _syscall/js_ than a standalone library. For some types, e.g. _any_ is converted into _js.Value_.

## Status/TODO

* Dividing into multplie packages is missing. Currently everything is created into a single package.
* Method/Enum rename - transformation support to get a better API.
* Union support are missing as methods depending of this is unusable.

### Go 1.12

* Change from js.Value to js.Wrapper?
* Callback changed to Func + return value

## Type conversion

The following will be applied to different WebIDL types:

### any

Any type is currently handled converted into a _js.Value_.

### callback

A function is generated with conversion method.

> TODO: Move conversion Go->WASM from implementation to public method.

```webidl
callback Foo = void (int a, int b);
```

```golang
// FooFromJS is converted a returned js.Value into a function that can be invoked.
func FooFromJS(_value js.Value) FunctionStringCallback {
```

### dictionary

Will generate a structure with corresponding field. When convered to/from _js.Value_, values are copied into a new javascript object.

> TODO: required values

### enum

A WebIDL enum is transformed into a Go enum.

Input

```webidl

enum Foo {
    "hello",
    "world"
};
```

Will be turned into following:

```golang
type Foo int

const (
    Hello Foo = iota
    World
)

```

### interface

The most used type in WebIDL. Generate a struct.

|Annotation   |Desciption|
|-------------|----------|
|NoGlobalScope|Generate an interface without a struct|

#### constant

Any constants are converted into a Go _const_ value.

#### attribute

For every attribute, a get and set method is generated. For read only attributes only a getter is created.

```webidl
interface Foo {
    attribute int bar;
};
```

#### method

A Go method or function is created for every method, depending if it's static or not. The method is trying to take care most of the conversion code.

```webidl
interface Foo {
    int bar(int a, int b);
};
```

### sequence

For types that can be used as a _js.TypeArray_, a _js.Value_ is used as method input type. Other sequence types are converted part of method invoke.

### union

WebIDL keyword _or_ can be used to define multiple input or output values that can be returned. It's like a very limitied _any_ type.

> TODO: unions are currently completey unusable. Any method or attribute that is depending on this union get a reference to an empty interface.

Example:

```webidl
typedef (DOMString or Function) TimerHandler;
```
