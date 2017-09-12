#include <unistd.h>
#include <iostream>

#include "eicio/reader.h"

void printUsage() {
    std::cerr << "Usage: eicio-summary [options] <input eicio file>\n";
    std::cerr << "options:\n";
    std::cerr << "  -g	decompress the stdin input with gzip\n";
    std::cerr << std::endl;
}

// http://programanddesign.com/cpp/human-readable-file-size-in-c/
char *readableByteSize(double size, char *buf) {
    int i = 0;
    const char *units[] = {"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"};
    while (size > 1024) {
        size /= 1024;
        i++;
    }
    sprintf(buf, "%.*f %s", i, size, units[i]);
    return buf;
}

int main(int argc, char **argv) {
    bool gzip = false;

    int opt;
    while ((opt = getopt(argc, argv, "gh")) != -1) {
        switch (opt) {
            case 'g':
                gzip = true;
                break;
            default:
                printUsage();
                exit(EXIT_FAILURE);
        }
    }

    std::string inputFilename;
    if (optind < argc) {
        inputFilename = argv[optind];
    } else {
        printUsage();
        exit(EXIT_FAILURE);
    }

    eicio::Reader *reader;
    if (inputFilename.compare("-") == 0)
        reader = new eicio::Reader(STDIN_FILENO, gzip);
    else
        reader = new eicio::Reader(inputFilename.c_str());

    unsigned long nEvents = 0;
    std::map<std::string, unsigned long> collBytes;
    std::map<unsigned long, bool> runs;

    while (auto eventHdr = reader->GetHeader()) {
        runs[eventHdr->runnumber()] = true;
        nEvents++;

        for (auto collHdr : eventHdr->payloadcollections()) {
            if (collBytes.find(collHdr.type()) == collBytes.end()) collBytes[collHdr.type()] == 0;
            collBytes[collHdr.type()] += collHdr.payloadsize();
        }

        delete eventHdr;
    }

    std::cout << "Number of runs: " << runs.size() << std::endl;
    std::cout << "Number of events: " << nEvents << std::endl;
    std::cout << "Total bytes for..." << std::endl;
    char buf[16];
    for (auto coll : collBytes)
        std::cout << "\t" << coll.first << ": " << readableByteSize(coll.second, buf) << std::endl;

    delete reader;
    return EXIT_SUCCESS;
}
