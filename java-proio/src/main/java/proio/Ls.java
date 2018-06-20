package proio;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import com.google.protobuf.ByteString;
import com.google.protobuf.Descriptors;
import com.google.protobuf.Message;

public class Ls {
  public static void main(String[] args) {
    if (args.length != 1) {
      System.out.println("Please provide one argument for the input file name");
      return;
    }

    try {
      Reader reader = new Reader(args[0]);

      int nEvents = 0;
      Map<String, ByteString> lastMetadata = new HashMap<String, ByteString>();
      for (Event event : reader) {
        Map<String, ByteString> metadata = event.getMetadata();
        for (Map.Entry<String, ByteString> entry : metadata.entrySet()) {
          if (lastMetadata.get(entry.getKey()) != entry.getValue()) {
            System.out.println("Metadata: " + entry.getKey() + ": " + entry.getValue());
          }
        }
        lastMetadata = metadata;
        System.out.println("EVENT: " + nEvents);
        for (String tag : event.getTags()) {
          System.out.println(tag);
          for (long id : event.getTaggedEntries(tag)) {
            Message entry = event.getEntry(id);
            System.out.println("-------------------");
            System.out.println(getMessageString(entry).replaceAll("\n", "\n\t"));
            System.out.println("-------------------");
          }
        }
        nEvents++;
      }

      reader.close();
    } catch (Throwable e) {
      e.printStackTrace();
    }
  }

  // Performs heavy lifting for collection entry introspection
  private static String getMessageString(Message msg) {
    String returnString = "";

    Descriptors.Descriptor desc = msg.getDescriptorForType();
    List<Descriptors.FieldDescriptor> fields = desc.getFields();

    for (Descriptors.FieldDescriptor field : fields) {
      if (!field.isRepeated()) {
        if (msg.hasField(field)) {
          returnString = returnString + "\n" + field.getName() + ": ";
          Object value = msg.getField(field);
          returnString = returnString + getFieldValueString(field, value);
        }
      } else {
        int count = msg.getRepeatedFieldCount(field);
        for (int i = 0; i < count; i++) {
          returnString = returnString + "\n" + field.getName() + "[" + Integer.toString(i) + "]: ";
          Object value = msg.getRepeatedField(field, i);
          String fieldString = getFieldValueString(field, value);
          returnString = returnString + fieldString;
        }
      }
    }

    return returnString;
  }

  // Performs heavy lifting for collection entry introspection
  private static String getFieldValueString(Descriptors.FieldDescriptor field, Object value) {
    String returnString = "";

    switch (field.getType()) {
      case INT32:
      case UINT32:
        returnString = Integer.toString((Integer) value);
        break;
      case INT64:
      case UINT64:
        returnString = Long.toString((Long) value);
        break;
      case FLOAT:
        returnString = Float.toString((Float) value);
        break;
      case DOUBLE:
        returnString = Double.toString((Double) value);
        break;
      case STRING:
        returnString = (String) value;
        break;
      case MESSAGE:
        returnString = getMessageString((Message) value).replaceAll("\n", "\n\t");
        break;
    }

    return returnString;
  }
}
