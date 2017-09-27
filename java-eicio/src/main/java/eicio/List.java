package eicio;

import com.google.protobuf.Message;

public class List
{
    public static void main( String[] args )
    {
		if (args.length != 1) {
			System.out.println("Please provide one argument");
			return;
		}

		try {
			Reader reader = null;
			reader = new Reader(args[0]);

			if (reader != null) {
					Event event = reader.get();
					event.get("MCParticle");
			}

			reader.close();
		} catch (Throwable e) {
			e.printStackTrace();
		}
    }
}
