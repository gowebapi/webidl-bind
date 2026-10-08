[Constructor(MutationCallback callback),
 Exposed=Window]
interface MutationObserver {
  void observe(Node target, optional MutationObserverInit options);
  void disconnect();
  sequence<MutationRecord> takeRecords();
};