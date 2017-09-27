package eicio;

import java.io.DataInputStream;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.InputStream;
import java.io.IOException;
import java.util.zip.GZIPInputStream;

import com.google.protobuf.InvalidProtocolBufferException;

public class Reader
{
	InputStream stream = null;
	InputStream subStream = null;

	public Reader(String filename) throws IOException {
		stream = new FileInputStream(filename);

		if (filename.endsWith(".gz")) {
			subStream = stream;
				this.stream = new GZIPInputStream(subStream);
		}
	}

	public Reader(InputStream stream, boolean compress) throws IOException {
		this.stream = stream;

		if (compress) {
			subStream = stream;
			this.stream = new GZIPInputStream(subStream);
		}
	}

	public Reader(InputStream stream) throws IOException {
		this(stream, false);
	}

	public void close() throws IOException {
		stream.close();

		if (subStream != null) {
			subStream.close();
		}
	}

	public Event get()
			throws IOException,
			InvalidProtocolBufferException {
		int n = this.syncToMagic();
		if (n < 4) return null;

		long headerSize = getUnsignedInt();
		long payloadSize = getUnsignedInt();

		byte[] headerBuf = new byte[(int)headerSize];
		int nRead = stream.read(headerBuf);
		if (nRead != headerSize) return null;
		Model.EventHeader header = Model.EventHeader.parseFrom(headerBuf);

		byte[] payload = new byte[(int)payloadSize];
		nRead = stream.read(payload);
		if (nRead != payloadSize) return null;
		
		Event event = new Event(header, payload);

		return event;
	}

	public Model.EventHeader getHeader()
			throws IOException,
			InvalidProtocolBufferException {
		int n = this.syncToMagic();
		if (n < 4) return null;

		long headerSize = getUnsignedInt();
		long payloadSize = getUnsignedInt();

		byte[] headerBuf = new byte[(int)headerSize];
		int nRead = stream.read(headerBuf);
		if (nRead != headerSize) return null;
		Model.EventHeader header = Model.EventHeader.parseFrom(headerBuf);

		stream.skip(payloadSize);

		return header;
	}

	private int syncToMagic() throws IOException {
		int nRead = 0;
		while (true) {
			int thisInt;
			if ((thisInt = stream.read()) == -1) break;
			byte thisByte = (byte)thisInt;
			nRead++;

			if (thisByte == Writer.magicBytes[0]) {
				boolean goodSeq = true;

				for (int i = 1; i < 4; i++) {
					if ((thisInt = stream.read()) == -1) break;
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

	private long getUnsignedInt() throws IOException {
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
