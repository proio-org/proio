#include <unistd.h>
#include <iostream>

#include "proio/reader.h"
#include "proio/writer.h"

void printUsage() {
    std::cerr << "Usage: proio-reserialize [options] <input proio file> <output proio file>\n";
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

    proio::Reader *reader;
    if (inputFilename.compare("-") == 0)
        reader = new proio::Reader(STDIN_FILENO, gzip);
    else
        reader = new proio::Reader(inputFilename.c_str());

    proio::Writer *writer;
    if (outputFilename.compare("-") == 0)
        writer = new proio::Writer(STDOUT_FILENO, compress);
    else
        writer = new proio::Writer(outputFilename.c_str());

    while (auto event = reader->Get()) {
        writer->Push(event);

        delete event;
    }

    delete writer, reader;
    return EXIT_SUCCESS;
}
