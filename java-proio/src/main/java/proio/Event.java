package proio;

import java.lang.ClassNotFoundException;
import java.lang.IllegalAccessException;
import java.lang.InstantiationException;
import java.lang.NoSuchMethodException;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.Arrays;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.HashMap;

import com.google.common.primitives.Bytes;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Descriptors;
import com.google.protobuf.DynamicMessage;
import com.google.protobuf.Message;
import com.google.protobuf.Parser;

public class Event
{
	public Model.EventHeader header = null;
	private byte[] payload = null;
	private Map<String, Message> collCache = new HashMap<String, Message>();

	public Event(Model.EventHeader header, byte[] payload) {
		this.header = header;
		this.payload = payload;
	}

	public List<String> getNames() {
		List<String> names = new ArrayList<String>();

		for (Model.EventHeader.CollectionHeader collHdr : header.getPayloadCollectionsList())
			names.add(collHdr.getName());
		for (String collName : collCache.keySet())
			names.add(collName);

		return names;
	}

	public Message get(String name)
			throws ClassNotFoundException,
			InstantiationException,
			InvalidProtocolBufferException,
			IllegalAccessException {
		Message coll = collCache.get(name);
		if (coll != null) return coll;

		return getFromPayload(name, true);
	}

	public Message dereference(Model.Reference ref)
			throws ClassNotFoundException,
			InstantiationException,
			InvalidProtocolBufferException,
			IllegalAccessException {
		Message refColl = null;
		for (Message coll : collCache.values()) {
			Descriptors.Descriptor desc = coll.getDescriptorForType();
			Descriptors.FieldDescriptor idFieldDesc = desc.findFieldByName("id");
			if (idFieldDesc == null) continue;
			int collID = (Integer)coll.getField(idFieldDesc);

			if (collID == ref.getCollID()) {
				refColl = coll;
				if (ref.getEntryID() == 0) return refColl;
				break;
			}
		}
		if (refColl == null) {
			for (Model.EventHeader.CollectionHeader collHdr : header.getPayloadCollectionsList()) {
				if (collHdr.getId() == ref.getCollID()) {
					refColl = get(collHdr.getName());
					if (ref.getEntryID() == 0) return refColl;
					break;
				}
			}
		}
		if (refColl == null) return null;

		Descriptors.Descriptor desc = refColl.getDescriptorForType();
		Descriptors.FieldDescriptor entriesFieldDesc = desc.findFieldByName("entries");
		if (entriesFieldDesc == null) return null;

		int entryCount = refColl.getRepeatedFieldCount(entriesFieldDesc);
		for (int i = 0; i < entryCount; i++) {
			Message entry = (Message)refColl.getRepeatedField(entriesFieldDesc, i);

			Descriptors.Descriptor entryDesc = entry.getDescriptorForType();
			Descriptors.FieldDescriptor entryIDFieldDesc = entryDesc.findFieldByName("id");
			if (entryIDFieldDesc == null) continue;
			int entryID = (Integer)entry.getField(entryIDFieldDesc);

			if (entryID == ref.getEntryID())
				return entry;
		}

		return null;
	}

	private Message getFromPayload(String name, boolean unmarshal)
			throws ClassNotFoundException,
			InstantiationException,
			InvalidProtocolBufferException,
			IllegalAccessException {
		int offset = 0;
		int size = 0;
		String collType = "";
		int collIndex = 0;
		Model.EventHeader.CollectionHeader collHdr = null;
		List collHdrList = header.getPayloadCollectionsList();
		for (collIndex = 0; collIndex < collHdrList.size(); collIndex++) {
			collHdr = (Model.EventHeader.CollectionHeader)collHdrList.get(collIndex);
			if (collHdr.getName().equals(name)) {
				collType = collHdr.getType();
				size = collHdr.getPayloadSize();
				break;
			}
			offset += collHdr.getPayloadSize();
		}
		if (collType == "")
			return null;

		Message coll = null;
		if (unmarshal) {
			Class collClass = Class.forName("proio.Model$" + collType);
			try {
				Method newBuilder = collClass.getMethod("newBuilder");
				Message.Builder builder = (Message.Builder)newBuilder.invoke(null);

				byte[] subPayload = Arrays.copyOfRange(payload, offset, offset + size);
				coll = builder.mergeFrom(subPayload).build();

				collCache.put(name, coll);
			} catch (NoSuchMethodException e) {
				e.printStackTrace();
			} catch (InvocationTargetException e) {
				e.printStackTrace();
			}
		}

		Model.EventHeader.Builder hdrBuilder = header.toBuilder();
		header = hdrBuilder.removePayloadCollections(collIndex).build();
		payload = Bytes.concat(Arrays.copyOfRange(payload, 0, offset), Arrays.copyOfRange(payload, offset + size, payload.length));

		return coll;
	}
}
