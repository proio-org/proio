package eicio;

import java.io.DataInputStream;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.InputStream;
import java.io.IOException;
import java.lang.Iterable;
import java.util.Iterator;
import java.util.zip.GZIPInputStream;

import com.google.protobuf.InvalidProtocolBufferException;

public class Reader implements Iterable<Event>, Iterator<Event>
{
	InputStream stream = null;
	InputStream subStream = null;

	private Event queuedEvent = null;

	public Reader(String filename)
			throws IOException {
		stream = new FileInputStream(filename);

		if (filename.endsWith(".gz")) {
			subStream = stream;
			this.stream = new GZIPInputStream(subStream);
		}
	}

	public Reader(InputStream stream, boolean compress)
			throws IOException {
		this.stream = stream;

		if (compress) {
			subStream = stream;
			this.stream = new GZIPInputStream(subStream);
		}
	}

	public Reader(InputStream stream)
			throws IOException {
		this(stream, false);
	}

	public void close()
			throws IOException {
		stream.close();

		if (subStream != null) {
			subStream.close();
		}
	}

	public Event get()
			throws IOException,
			InvalidProtocolBufferException {
		if (queuedEvent != null) {
			Event returnEvent = queuedEvent;
			queuedEvent = null;
			return returnEvent;
		}

		int n = this.syncToMagic();
		if (n < 4) return null;

		long headerSize = getUnsignedInt();
		long payloadSize = getUnsignedInt();

		byte[] headerBuf = new byte[(int)headerSize];
		int nRead = getBytes(headerBuf);
		if (nRead != headerSize) return null;
		Model.EventHeader header = Model.EventHeader.parseFrom(headerBuf);

		byte[] payload = new byte[(int)payloadSize];
		nRead = getBytes(payload);
		if (nRead != payloadSize) return null;
		
		Event event = new Event(header, payload);

		return event;
	}

	public Model.EventHeader getHeader()
			throws IOException,
			InvalidProtocolBufferException {
		if (queuedEvent != null) {
			Event returnEvent = queuedEvent;
			queuedEvent = null;
			return returnEvent.header;
		}

		int n = this.syncToMagic();
		if (n < 4) return null;

		long headerSize = getUnsignedInt();
		long payloadSize = getUnsignedInt();

		byte[] headerBuf = new byte[(int)headerSize];
		int nRead = getBytes(headerBuf);
		if (nRead != headerSize) return null;
		Model.EventHeader header = Model.EventHeader.parseFrom(headerBuf);

		skipBytes(payloadSize);

		return header;
	}

	public int skip(int nEvents)
			throws IOException {
		int nSkipped = 0;
		for (int i = 0; i < nEvents; i++) {
			if (queuedEvent != null) {
				queuedEvent = null;
				nSkipped++;
				continue;
			}

			int n = this.syncToMagic();
			if (n < 4) return nSkipped;

			long headerSize = getUnsignedInt();
			long payloadSize = getUnsignedInt();

			skipBytes(headerSize + payloadSize);
			nSkipped++;
		}

		return nSkipped;
	}

	public Iterator<Event> iterator() {
		return this;
	}

	public boolean hasNext() {
		try {
			if ((queuedEvent = get()) != null)
				return true;
		} catch (Throwable e) {
			e.printStackTrace();
		}
		return false;
	}

	public Event next() {
		try {
			return get();
		} catch (Throwable e) {
			e.printStackTrace();
		}
		return null;
	}

	private int getBytes(byte[] buf)
			throws IOException {
		int nRead = 0;
		while (true) {
			int nBytes = stream.read(buf, nRead, buf.length - nRead);
			if (nBytes < 0) return nRead;
			nRead += nBytes;
			if (nRead == buf.length) return nRead;
		}
	}

	private long skipBytes(long n)
			throws IOException {
		long nSkipped = 0;
		while (true) {
			long nBytes = stream.skip(n - nSkipped);
			if (nBytes < 0L) return nSkipped;
			nSkipped += nBytes;
			if (nSkipped == n) return nSkipped;
		}
	}

	private int syncToMagic()
			throws IOException {
		int nRead = 0;
		while (true) {
			int thisInt;
			if ((thisInt = stream.read()) == -1) return -1;
			byte thisByte = (byte)thisInt;
			nRead++;

			if (thisByte == Writer.magicBytes[0]) {
				boolean goodSeq = true;

				for (int i = 1; i < 4; i++) {
					if ((thisInt = stream.read()) == -1) return -1;
					thisByte = (byte)thisInt;
					nRead++;

					if (thisByte != Writer.magicBytes[i]) {
						goodSeq = false;
						break;
					}
				}

				if (goodSeq) break;
			}
		}

		return nRead;
	}

	private long getUnsignedInt()
			throws IOException {
		final int nBytes = 4;
		int[] theseBytes = new int[nBytes];
		for (int i = 0; i < nBytes; i++) {
			theseBytes[i] = stream.read();
			if (theseBytes[i] == -1) return -1L;
		}

		long num = 0;
		for (int i = 0; i < nBytes; i++)
			num |= (theseBytes[i] & 0xFFL) << (i * 8);
		return num;
	}
}
