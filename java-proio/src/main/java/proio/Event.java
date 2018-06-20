package proio;

import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import java.util.Vector;
import org.reflections.Reflections;

public class Event {
  protected Proto.Event eventProto = null;
  protected Map<String, ByteString> metadata = new HashMap<String, ByteString>();

  private Map<Long, Message> entryCache = new HashMap<Long, Message>();
  private Map<Long, Class> classCache = new HashMap<Long, Class>();
  private static Reflections refl = new Reflections("proio.model");

  public Event() {}

  public Iterable<Long> getAllEntries() {
    Vector<Long> entries = new Vector<Long>();
    if (eventProto != null) {
      for (Map.Entry<Long, Proto.Any> entry : eventProto.getEntryMap().entrySet()) {
        entries.add(entry.getKey());
      }
    }
    return entries;
  }

  public Iterable<Long> getTaggedEntries(String tag) {
    if (eventProto == null || !eventProto.containsTag(tag)) {
      return new Vector<Long>();
    }

    return eventProto.getTagMap().get(tag).getEntryList();
  }

  public Message getEntry(long id)
      throws NoSuchMethodException, IllegalAccessException, InvalidProtocolBufferException,
          ClassNotFoundException, InvocationTargetException {
    if (entryCache.containsKey(id)) {
      return entryCache.get(id);
    }

    if (!eventProto.getEntryMap().containsKey(id)) {
      return null;
    }
    Proto.Any entryProto = eventProto.getEntryMap().get(id);

    Class entryClass = getClass(entryProto.getType());
    Method newBuilder = entryClass.getMethod("newBuilder");
    Message.Builder builder = (Message.Builder) newBuilder.invoke(null);
    builder.mergeFrom(entryProto.getPayload());
    Message entry = builder.build();

    entryCache.put(id, entry);
    return entry;
  }

  public Iterable<String> getTags() {
    Vector<String> tags = new Vector<String>();
    if (eventProto != null) {
      for (Map.Entry<String, Proto.Tag> entry : eventProto.getTagMap().entrySet()) {
        tags.add(entry.getKey());
      }
    }
    Collections.sort(tags);
    return tags;
  }

  public Map<String, ByteString> getMetadata() {
    Map map = new HashMap<String, ByteString>();
    for (Map.Entry<String, ByteString> entry : metadata.entrySet()) {
      map.put(entry.getKey(), entry.getValue());
    }
    return map;
  }

  public void clear() {
    eventProto = null;
    metadata.clear();
    return;
  };

  private Class getClass(long typeID) throws ClassNotFoundException {
    if (!classCache.containsKey(typeID)) {
      String typeName = eventProto.getTypeMap().get(typeID).toLowerCase();
      for (Class thisClass : refl.getSubTypesOf(Message.class)) {
        String thisName = thisClass.getName().replace('$', '.').toLowerCase();
        if (thisName.equals(typeName)) {
          classCache.put(typeID, thisClass);
          return thisClass;
        }
      }
    }
    return classCache.get(typeID);
  }
}
