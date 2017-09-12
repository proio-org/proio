#include <unistd.h>
#include <iostream>

#include "eicio/reader.h"
#include "eicio/writer.h"

void printUsage() {
    std::cerr << "Usage: eicio-reserialize [options] <input eicio file> <output eicio file>\n";
    std::cerr << "options:\n";
    std::cerr << "  -g	decompress the stdin input with gzip\n";
    std::cerr << "  -c	compress the stdout output with gzip\n";
    std::cerr << std::endl;
}

int main(int argc, char **argv) {
    bool compress = false;
    bool gzip = false;

    int opt;
    while ((opt = getopt(argc, argv, "cgh")) != -1) {
        switch (opt) {
            case 'c':
                compress = true;
                break;
            case 'g':
                gzip = true;
                break;
            default:
                printUsage();
                exit(EXIT_FAILURE);
        }
    }

    std::string inputFilename;
    std::string outputFilename;
    if (optind + 1 < argc) {
        inputFilename = argv[optind];
        outputFilename = argv[optind + 1];
    } else {
        printUsage();
        exit(EXIT_FAILURE);
    }

    eicio::Reader *reader;
    if (inputFilename.compare("-") == 0)
        reader = new eicio::Reader(STDIN_FILENO, gzip);
    else
        reader = new eicio::Reader(inputFilename.c_str());

    eicio::Writer *writer;
    if (outputFilename.compare("-") == 0)
        writer = new eicio::Writer(STDOUT_FILENO, compress);
    else
        writer = new eicio::Writer(outputFilename.c_str());

    while (auto event = reader->Get()) {
        writer->Push(event);

        delete event;
    }

    delete writer, reader;
    return EXIT_SUCCESS;
}
