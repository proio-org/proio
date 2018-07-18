#undef NDEBUG

#include <assert.h>
#include <stdio.h>
#include <iostream>

#include "event.h"
#include "reader.h"
#include "writer.h"

void pushUpdate1() {
    char filename1[] = "pushupdate1.1XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename1));

    writer->PushMetadata("key1", "value1");
    writer->PushMetadata("key2", "value2");
    auto event = new proio::Event();
    writer->Push(event);
    writer->PushMetadata("key2", "value3");
    writer->Push(event);
    std::string value4 = "value4";
    std::string value5 = "value5";
    writer->PushMetadata("key1", value4);
    writer->PushMetadata("key2", value5);
    writer->Push(event);

    delete writer;
    delete event;
    auto reader = new proio::Reader(filename1);

    auto event1 = reader->Next();
    auto event2 = reader->Next();
    auto event3 = reader->Next();
    assert(event1->Metadata().at("key1")->compare("value1") == 0);
    assert(event1->Metadata().at("key2")->compare("value2") == 0);
    assert(event2->Metadata().at("key1")->compare("value1") == 0);
    assert(event2->Metadata().at("key2")->compare("value3") == 0);
    assert(event3->Metadata().at("key1")->compare("value4") == 0);
    assert(event3->Metadata().at("key2")->compare("value5") == 0);

    delete reader;
    char filename2[] = "pushupdate1.2XXXXXX";
    writer = new proio::Writer(mkstemp(filename2));

    writer->Push(event1);
    writer->Push(event2);
    writer->Push(event3);

    delete event1;
    delete event2;
    delete event3;
    delete writer;
    reader = new proio::Reader(filename1);

    event1 = reader->Next();
    event2 = reader->Next();
    event3 = reader->Next();
    assert(event1->Metadata().at("key1")->compare("value1") == 0);
    assert(event1->Metadata().at("key2")->compare("value2") == 0);
    assert(event2->Metadata().at("key1")->compare("value1") == 0);
    assert(event2->Metadata().at("key2")->compare("value3") == 0);
    assert(event3->Metadata().at("key1")->compare("value4") == 0);
    assert(event3->Metadata().at("key2")->compare("value5") == 0);

    delete event1;
    delete event2;
    delete event3;
    delete reader;
    remove(filename1);
    remove(filename2);
}

void pushUpdate2() {
    auto event = new proio::Event();
    char filename1[] = "pushupdate2.1XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename1));

    writer->PushMetadata("key1", "value1");
    writer->PushMetadata("key2", "value2");
    writer->Push(event);
    writer->PushMetadata("key2", "value3");
    writer->Push(event);
    std::string value4 = "value4";
    std::string value5 = "value5";
    writer->PushMetadata("key1", value4);
    writer->PushMetadata("key2", value5);
    writer->Push(event);

    delete writer;
    auto reader = new proio::Reader(filename1);
    char filename2[] = "pushupdate2.2XXXXXX";
    writer = new proio::Writer(mkstemp(filename2));

    reader->Next(event, true);
    assert(event->Metadata().at("key1")->compare("value1") == 0);
    assert(event->Metadata().at("key2")->compare("value2") == 0);
    writer->Push(event);
    reader->Next(event, true);
    assert(event->Metadata().at("key1")->compare("value1") == 0);
    assert(event->Metadata().at("key2")->compare("value3") == 0);
    writer->Push(event);
    reader->Next(event, true);
    assert(event->Metadata().at("key1")->compare("value4") == 0);
    assert(event->Metadata().at("key2")->compare("value5") == 0);
    writer->Push(event);

    delete writer;
    delete reader;
    reader = new proio::Reader(filename1);

    reader->Next(event, true);
    assert(event->Metadata().at("key1")->compare("value1") == 0);
    assert(event->Metadata().at("key2")->compare("value2") == 0);
    reader->Next(event, true);
    assert(event->Metadata().at("key1")->compare("value1") == 0);
    assert(event->Metadata().at("key2")->compare("value3") == 0);
    reader->Next(event, true);
    assert(event->Metadata().at("key1")->compare("value4") == 0);
    assert(event->Metadata().at("key2")->compare("value5") == 0);

    delete reader;
    remove(filename1);
    remove(filename2);
    delete event;
}

void pushUpdateRestart1() {
    char filename[] = "pushupdaterestart1XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));

    writer->PushMetadata("key1", "value1");
    writer->PushMetadata("key2", "value2");
    auto event = new proio::Event();
    writer->Push(event);
    writer->PushMetadata("key2", "value3");
    writer->Push(event);
    std::string value4 = "value4";
    std::string value5 = "value5";
    std::string value6 = "value6";
    writer->PushMetadata("key1", value4);
    writer->PushMetadata("key2", value5);
    writer->PushMetadata("key3", value6);
    writer->Push(event);

    delete writer;
    delete event;
    auto reader = new proio::Reader(filename);

    delete reader->Next();
    delete reader->Next();
    delete reader->Next();
    reader->SeekToStart();

    auto event1 = reader->Next();
    auto event2 = reader->Next();
    auto event3 = reader->Next();
    assert(event1->Metadata().at("key1")->compare("value1") == 0);
    assert(event1->Metadata().at("key2")->compare("value2") == 0);
    assert(event1->Metadata().count("key3") == 0);
    assert(event2->Metadata().at("key1")->compare("value1") == 0);
    assert(event2->Metadata().at("key2")->compare("value3") == 0);
    assert(event2->Metadata().count("key3") == 0);
    assert(event3->Metadata().at("key1")->compare("value4") == 0);
    assert(event3->Metadata().at("key2")->compare("value5") == 0);
    assert(event3->Metadata().at("key3")->compare("value6") == 0);

    delete event1;
    delete event2;
    delete event3;
    delete reader;
    remove(filename);
}
int main() {
    pushUpdate1();
    pushUpdate2();
    pushUpdateRestart1();
}
