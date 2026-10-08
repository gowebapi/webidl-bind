

interface Foo1 {
    maplike<DOMString, unsigned long>;
};

// extension, not defined in webidl spec
interface mixin Foo2 {
    maplike<DOMString, unsigned long>;
};

interface Foo3 {
    readonly maplike<DOMString, unsigned long>;
};

// extension, not defined in webidl spec
interface mixin Foo4 {
    readonly maplike<DOMString, unsigned long>;
};

interface Foo5 {
    setlike<DOMString>;
};

interface Foo6 {
    readonly setlike<unsigned long>;
};

interface Foo7 {
    iterable<long>;
}

interface Foo8 {
    iterable<long, DOMString>;
};

interface Foo9 {
    async iterable<long>;
};

interface Foo10 {
    async iterable<long, DOMString>;
};
