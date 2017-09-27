package eicio;

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

	public Message get(String name)
			throws ClassNotFoundException,
			InstantiationException,
			InvalidProtocolBufferException,
			IllegalAccessException {
		Message coll = collCache.get(name);
		if (coll != null) return coll;

		return getFromPayload(name, true);
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
			Class collClass = Class.forName("eicio.Model$" + collType);
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
