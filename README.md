
# Code binding generator for Web Assembly from WebIDL

This is a code generator that is taking WebIDL file and create binding code.
It can currently generate DOM/HTML bindings for Go Web Assembly.

## Target Languages

* Go 1.11 Web Assembly - Initail support

Current no other languages is planned. Contributions are welcome :)

## Input files

The hardest part is not to write a WebIDL from a specification, but keeping it up to date. The philosophy is extract webidl from a third source, e.g. by taking all IDL part from the DOM specification (<https://dom.spec.whatwg.org)> and having all the modifications in other files.

The program need following input files:

* foo.idl - Main WebIDL file, a "1:1" copy of the specification.
* foo.addition.idl - Extra types and oddity that exist in the specification. E.g. the SVG IDL are refering to DOMRect that doesn't exist anywhere.
* foo.go.md - Language transformation file. To modify incoming WebIDL and turning it into something thats look like a "standard library"
* foo.doc.md - (Planned) API documentation file

## Status/TODO

Currently the generator can process the DOM and HTML specification and create a compilable output. There are still missing feature, see [Go WASM](gowasm.md) for details.

## WebIDL

WebIDL specification can be found at <https://heycam.github.io/webidl/>

### Global scope

The specification files doesn't containts browsers global varibale scope, e.g. access to _window_. This can be defined with a special annotation _OnGlobalScope_ on a interface be able to define this methods and attributes. Please note that all attributes and methods need to be defined static to get correctly compilable code.

```webidl
[OnGlobalScope]
interface GlobalScope {
    // access to javascript document variable.
    static readonly attribute Document document;
    // access to javascript window variable.
    static readonly attribute Window window;
};
```

> Note: in the above example, to generator will create a function named Document() to get the document attribute. This will name clash with the interface Document. This is fixed by the language transformation file that is renaming the attribute in the final lanaguage.

## Language transformation file

The transformation file are used to fix issues to get a final output that feels more "natrual" than working with raw generated files. Examples:

* Method and constant rename. Go doesn't have support for static methods and constants , javascript does. Any static methods writted outside of the structure and this can leed to name clash if two interfaces define the same static method.
* Enum value rename, e.g. by turning notification api "ltr" to LeftToRight.
* Move interfaces to other packages. DOM and HTML specification have cirular dependency between them, some interfaces/methods need to be moved to get a compilable output.

Current format using MarkDown ending to get some IDE syntax highlightning. With the exception for header tags (##), no other MarkDown synta is supported.

```markdown

# Initail header have no meaning

    Any line starting with tab or spaces is consider to be a comment line

## Foo ("WebIDL type name")

    any line starting with a dot is modificing properties on type it self, e.g. rename the type to Bar

.name = Bar

    any other lines with equal sign is renaming method or attributes to target lanaguage name.
methodName = languageName

    Developers need to invoke SayHelloWorld() in target language to trigger helloWorld() in javascript.
helloWorld = SayHelloWorld

```

### Callback

|Syntax Name|Description|Default|
|-----------|-----------|-------|
|.package|package name|first part of the input file|
|.name|type output name|idl type name in public access format|

### Dictionary

|Syntax Name|Description|Default|
|-----------|-----------|-------|
|.package|package name|first part of the input file|
|.name|type output name|idl type name in public access format|

### Enum

|Syntax Name|Description|Default|
|-----------|-----------|-------|
|.package|package name|first part of the input file|
|.name|type output name|idl type name in public access format|

### Interface

Interfaces have following properites

|Syntax Name|Description|Default|
|-----------|-----------|-------|
|.package|package name|first part of the input file|
|.name|type output name|idl type name in public access format|
|.constPrefix|a prefix added to all type constants|empty|
|.constSuffix|a suffix added to all type constants|interface name|
