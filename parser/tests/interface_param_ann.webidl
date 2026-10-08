[Exposed=Window]
interface DOMImplementation {
  [NewObject] DocumentType createDocumentType(DOMString qualifiedName, DOMString publicId, DOMString systemId);
  [NewObject] XMLDocument createDocument(DOMString? namespace, [TreatNullAs=EmptyString] DOMString qualifiedName, optional DocumentType? doctype = null);
  [NewObject] Document createHTMLDocument(optional DOMString title);

  boolean hasFeature(); // useless; always returns true
};

[Exposed=Window]
interface CharacterData : Node {
  attribute [TreatNullAs=EmptyString] DOMString data;
  readonly attribute unsigned long length;
  DOMString substringData(unsigned long offset, unsigned long count);
  void appendData(DOMString data);
  void insertData(unsigned long offset, DOMString data);
  void deleteData(unsigned long offset, unsigned long count);
  void replaceData(unsigned long offset, unsigned long count, DOMString data);
  void postMessage(any message, optional sequence<object> transfer = []);
};