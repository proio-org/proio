package eicio;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import com.google.protobuf.Descriptors;
import com.google.protobuf.Message;

public class Ls
{
    public static void main( String[] args )
    {
		if (args.length != 1) {
			System.out.println("Please provide one argument for the input file name");
			return;
		}

		try {
			Reader reader = new Reader(args[0]);

			if (reader == null) return;

			for (Event event : reader) {
				for (String collName : event.getNames()) {
					Message coll = event.get(collName);
					Descriptors.Descriptor desc = coll.getDescriptorForType();

					Descriptors.FieldDescriptor idFieldDesc = desc.findFieldByName("id");
					Descriptors.FieldDescriptor flagsFieldDesc = desc.findFieldByName("flags");
					Descriptors.FieldDescriptor paramsFieldDesc = desc.findFieldByName("params");
					Descriptors.FieldDescriptor entriesFieldDesc = desc.findFieldByName("entries");

					int collID = (Integer)coll.getField(idFieldDesc);
					int collFlags = (Integer)coll.getField(flagsFieldDesc);
					Model.Params collParams = (Model.Params)coll.getField(paramsFieldDesc);

					System.out.println("collName: " + collName);
					System.out.println("collID: " + Integer.toString(collID));
					System.out.println("collFlags: " + Integer.toString(collFlags));
					System.out.println("collParams:" + getMessageString(collParams).replaceAll("\n", "\n\t"));

					int nEntries = coll.getRepeatedFieldCount(entriesFieldDesc);
					System.out.print("entries (" + Integer.toString(nEntries) + "):");
					for (int i = 0; i < nEntries; i++) {
						Message entry = (Message)coll.getRepeatedField(entriesFieldDesc, i);
						System.out.println(getMessageString(entry).replaceAll("\n", "\n\t"));
					}
				}
			}

			reader.close();
		} catch (Throwable e) {
			e.printStackTrace();
		}
    }

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

	private static String getFieldValueString(Descriptors.FieldDescriptor field, Object value) {
		String returnString = "";

		switch (field.getType()) {
			case INT32:
			case UINT32:
				returnString = Integer.toString((Integer)value);
				break;
			case INT64:
			case UINT64:
				returnString = Long.toString((Long)value);
				break;
			case FLOAT:
				returnString = Float.toString((Float)value);
				break;
			case DOUBLE:
				returnString = Double.toString((Double)value);
				break;
			case STRING:
				returnString = (String)value;
				break;
			case MESSAGE:
				returnString = getMessageString((Message)value).replaceAll("\n", "\n\t");
				break;
		}

		return returnString;
	}
}
