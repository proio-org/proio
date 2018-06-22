#undef NDEBUG

#include <assert.h>

#include "eic.pb.h"
#include "event.h"
#include "reader.h"
#include "writer.h"

using namespace proio::model;

void serialize1() {
    auto event1 = new proio::Event();
    event1->AddEntry(new eic::Particle(), "Particle");
    event1->AddEntry(new eic::Particle(), "Particle");
    std::string data;
    event1->SerializeToString(&data);

    auto event2 = new proio::Event();
    assert(event1->String().compare(event2->String()) != 0);
    delete event2;

    auto event3 = new proio::Event(data);
    assert(event1->String().compare(event3->String()) == 0);
    delete event3;

    delete event1;
}

void serialize2(proio::Compression comp) {
    char filename[] = "serialize2XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));
    writer->SetCompression(comp);

    auto event = new proio::Event();
    event->AddEntry(new eic::Particle(), "Particle");
    event->AddEntry(new eic::Particle(), "Particle");
    std::string data1;
    event->SerializeToString(&data1);
    writer->Push(event);
    delete event;

    delete writer;

    auto reader = new proio::Reader(filename);
    std::string data2;
    reader->Next(&data2);
    assert(data1.compare(data2) == 0);

    remove(filename);
}

int main() {
    serialize1();
    serialize2(proio::LZ4);
    serialize2(proio::UNCOMPRESSED);
    serialize2(proio::GZIP);
}
