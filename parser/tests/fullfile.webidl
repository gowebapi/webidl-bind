// From: https://groups.google.com/a/chromium.org/forum/#!topic/blink-dev/KPpVpf7Mv8k

[NoInterfaceObject]
interface ECMA262Globals {
  readonly attribute Number NaN;
  readonly attribute Number Infinity;
  readonly attribute Object undefined;

  any eval(any x);
  int parseInt(String string, int radix);
  Number parseFloat(String string);
  boolean isNaN(Number number);
  boolean isFinite(Number number);
  String decodeURI(any encodedUri);
  String decodeURIComponent(any encodedUri);
  String encodeURI(any uri);
  String encodeURIComponent(any uri);

};
Window implements ECMA262Globals;

[Constructor]
interface Object {
  static readonly attribute Object prototype;
  static Object getPrototypeOf(Object object);
  static Object getOwnPropertyDescriptor(Object object, any name);
  static Array getOwnPropertyNames(Object object);
  static Object create(Object object, optional any properties);
  static Object defineProperty(Object object, any name, Object attributes);
  static Object defineProperties(Object object, Object attributes);
  static Object seal(Object object);
  static Object freeze(Object object);
  static Object preventExtensions(Object object);
  static boolean isSealed(Object object);
  static boolean isFrozen(Object object);
  static boolean isExtensible(Object object);
  static Array keys(Object object);
  String toString();
  String toLocaleString();
  Object valueOf();
  boolean hasOwnProperty(any value);
  boolean isPrototypeOf(any value);
  boolean propertyIsEnumerable(any value);
};

[Constructor]
interface Function {
  static readonly attribute Object prototype;
  static readonly attribute int length;
  String toString();
  any apply(Object this, Array arguments);
  any call(Object this, optional any item1);
  any bind(Object this, optional any item1);
  readonly attribute Object prototype;
  readonly attribute int length;
};

[Constructor]
interface Array {
  static readonly attribute Object prototype;
  static boolean isArray(any x);
  String toString();
  String toLocaleString();
  Array concat(optional any item1);
  String join(any separator);
  Array pop();
  Array push(optional any item1);
  Array reverse();
  any shift();
  Array slice(any start, any end);
  Array sort(any comparator);
  Array splice(any start, any deleteCount, optional any item1);
  int unshift(optional any item1);
  int indexOf(any searchElement, optional int fromIndex);
  int lastIndexOf(any searchElement, optional int fromIndex);
  boolean every(any predicator, Object this);
  boolean some(any predicator, Object this);
  void forEach(any iterator, Object this);
  Array map(any iterator, Object this);
  Array filter(any predicator, Object this);
  any reduce(any reductor, optional any initial);
  any reduceRight(any reductor, optional any initial);
  attribute int length;
};

[Constructor]
interface String {
  static readonly attribute Object prototype;
  static String fromCharCode(optional Number char1);
  String toString();
  String valueOf();
  String charAt(any position);
  Number charCodeAt(any position);
  String concat(optional any string1);
  int indexOf(optional any searchString, optional any position);
  int lastIndexOf(optional any searchString, optional any position);
  Number localeCompare(any that);
  Array match(RegExp pattern);
  String replace(any pattern, any replacement);
  Number search(RegExp pattern);
  String slice(any start, any end);
  Array split(any separator, optional Number limit);
  String substring(any start, any end);
  String toLowerCase();
  String toLocaleLowerCase();
  String toUpperCase();
  String toLocaleUpperCase();
  String trim();
  attribute int length;
};

[Constructor(any value)]
interface Boolean {
  static readonly attribute Object prototype;
  String toString();
  Boolean valueOf();
};

[Constructor(optional any value)]
interface Number {
  static readonly attribute Object prototype;
  static readonly attribute Number MAX_VALUE;
  static readonly attribute Number MIN_VALUE;
  static readonly attribute Number NaN;
  static readonly attribute Number NEGATIVE_INFINITY;
  static readonly attribute Number POSITIVE_INFINITY;
  String toString(optional any radix);
  String toLocaleString();
  Number valueOf();
  String toFixed(any fractionDigits);
  String toExponential(any fractionDigits);
  String toPrecision(any precision);
};

[NoInterfaceObject]
interface Math {
  readonly attribute Number E;
  readonly attribute Number LN10;
  readonly attribute Number LN2;
  readonly attribute Number LOG2E;
  readonly attribute Number LOG10E;
  readonly attribute Number PI;
  readonly attribute Number SQRT1_2;
  readonly attribute Number SQRT2;
  Number abs(Number x);
  Number acos(Number x);
  Number asin(Number x);
  Number atan(Number x);
  Number atan2(Number x, Number y);
  Number ceil(Number x);
  Number cos(Number x);
  Number exp(Number x);
  Number floor(Number x);
  Number log(Number x);
  Number max(optional Number value1);
  Number min(optional Number value1);
  Number pow(Number x, Number y);
  Number random();
  Number round(Number x);
  Number sin(Number x);
  Number sqrt(Number x);
  Number tan(Number x);
};

[
 Constructor,
 Constructor(any value),
 Constructor(any year, any month, optional any date, optional any hour, optional any minute, optional any second, optional any ms)
]
interface Date {
  static readonly attribute Object prototype;
  static Number parse(any string);
  static Number UTC(any year, any month, optional any date, optional any hour, optional any minute, optional any second, optional any ms);
  static Number now();
  String toString();
  String toDateString();
  String toTimeString();
  String toLocaleString();
  String toLocaleDateString();
  String toLocaleTimeString();
  Number valueOf();
  Number getTime();
  Number getFullYear();
  Number getUTCFullYear();
  Number getMonth();
  Number getUTCMonth();
  Number getDate();
  Number getUTCDate();
  Number getDay();
  Number getUTCDay();
  Number getHours();
  Number getUTCHours();
  Number getMinutes();
  Number getUTCMinutes();
  Number getSeconds();
  Number getUTCSeconds();
  Number getMilliseconds();
  Number getUTCMilliseconds();
  Number getTimezoneOffset();
  Number setTime(any time);
  Number setMilliseconds(any ms);
  Number setUTCMilliseconds(any ms);
  Number setSeconds(any seconds, optional any ms);
  Number setUTCSeconds(any seconds, optional any ms);
  Number setMinutes(any minutes, optional any seconds, optional any ms);
  Number setUTCMinutes(any minutes, optional any seconds, optional any ms);
  Number setHours(any hours, optional any minutes, optional any seconds, optional any ms);
  Number setUTCHours(any hours, optional any minutes, optional any seconds, optional any ms);
  Number setDate(any date);
  Number setUTCDate(any date);
  Number setMonth(any month, optional any date);
  Number setUTCMonth(any month, optional any date);
  Number setFullYear(any year, optional any month, optional any date);
  Number setUTCFullYear(any  year, optional any month, optional any date);
  String toUTCString();
  String toISOString();
  String toJSON(any key);
};

[Constructor(any pattern, any flags)]
interface RegExp {
  static readonly attribute Object prototype;
  readonly attribute String source;
  readonly attribute Boolean global;
  readonly attribute Boolean ignoreCase;
  readonly attribute Boolean multiline;
  attribute Number lastIndex;
  Array exec(any string);
  boolean test(any string);
  String toString();
};

[Constructor(any message)]
interface Error {
  static readonly attribute Object prototype;
  readonly attribute String name;
  readonly attribute String message;
  String toString();
};

[Constructor(any message)]
interface EvalError {
  static readonly attribute Object prototype;
  readonly attribute String name;
  readonly attribute String message;
  String toString();
};

[Constructor(any message)]
interface RangeError {
  static readonly attribute Object prototype;
  readonly attribute String name;
  readonly attribute String message;
  String toString();
};

[Constructor(any message)]
interface ReferenceError {
  static readonly attribute Object prototype;
  readonly attribute String name;
  readonly attribute String message;
  String toString();
};

[Constructor(any message)]
interface SyntaxError {
  static readonly attribute Object prototype;
  readonly attribute String name;
  readonly attribute String message;
  String toString();
};

[Constructor(any message)]
interface TypeError {
  static readonly attribute Object prototype;
  readonly attribute String name;
  readonly attribute String message;
  String toString();
};

[Constructor(any message)]
interface URIError {
  static readonly attribute Object prototype;
  readonly attribute String name;
  readonly attribute String message;
  String toString();
};

[NoInterfaceObject]
interface JSON {
  Object parse(any text, optional Function reviver);
  String stringify(any value, optional any replacer, optional any space);
};

