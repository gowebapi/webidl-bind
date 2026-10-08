[Global]
[Constructor()]
interface XMLHttpRequest {
	attribute Number status;
	attribute String statusText;
	attribute String responseText;
	attribute Number readyState;

	void send(any body);
	void open(String method, String url);
	void addEventListener(String eventName, any listener);
};