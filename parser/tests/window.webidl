interface HTMLDocument {
	
};

interface HTMLWindow {
   readonly attribute HTMLDocument document;
};

[Global]
interface Window {
   static readonly attribute HTMLDocument document;
};