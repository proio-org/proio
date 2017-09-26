package eicio;

import java.io.IOException;

public class List
{
    public static void main( String[] args )
    {
		if (args.length != 1) {
			System.out.println("Please provide one argument");
			return;
		}

		Reader reader = null;
		try {
			reader = new Reader(args[0]);
		} catch (IOException e) {
			e.printStackTrace();
		}

		if (reader != null) {
			try {
				reader.get();
			} catch (IOException e) {
				e.printStackTrace();
			}
		}

		try {
			reader.close();
		} catch (IOException e) {
			e.printStackTrace();
		}
    }
}
