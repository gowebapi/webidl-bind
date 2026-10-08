interface Element {
	readonly attribute String tagName;
	attribute String id;
	attribute String className;
	String? getAttribute(String name);
	void setAttribute(String name, String value);
	boolean hasAttribute(String name);
};

interface Document {
  Element createElement(String localName);
  Text createTextNode(String data);
};

[GlobalContext]
interface Window {
   readonly attribute Document document;
};