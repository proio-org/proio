package eicio;

import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.InputStream;
import java.io.IOException;
import java.util.zip.GZIPInputStream;

import com.google.protobuf.Message;

public class Reader
{
	InputStream stream = null;
	InputStream subStream = null;

	public Reader(String filename) throws IOException {
		try {
			stream = new FileInputStream(filename);
		} catch (FileNotFoundException e) {
			throw new IOException("Failure to open input file", e);
		}

		if (filename.endsWith(".gz")) {
			subStream = stream;
			try {
				this.stream = new GZIPInputStream(subStream);
			} catch (IOException e) {
				throw new IOException("Failure to create gzip stream", e);
			}
		}
	}

	public Reader(InputStream stream, boolean compress) throws IOException {
		this.stream = stream;

		if (compress) {
			subStream = stream;
			try {
				this.stream = new GZIPInputStream(subStream);
			} catch (IOException e) {
				throw new IOException("Failure to create gzip stream", e);
			}
		}
	}

	public Reader(InputStream stream) throws IOException {
		this(stream, false);
	}

	public void close() throws IOException {
		try {
			stream.close();
		} catch (IOException e) {
			throw new IOException("Failure to close stream", e);
		}

		if (subStream != null) {
			try {
				subStream.close();
			} catch (IOException e) {
				throw new IOException("Failure to close substream", e);
			}
		}
	}

	public Message get() throws IOException {
		int n = this.syncToMagic();
		return null;
	}

	private int syncToMagic() throws IOException {
		for (int i = 0; i < 4; i++) {
			try {
				System.out.println(stream.read());
			} catch (IOException e) {
				throw new IOException("Failure to read magic number", e);
			}
		}

		return 0;
	}
}
