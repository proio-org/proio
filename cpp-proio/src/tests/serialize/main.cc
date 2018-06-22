#undef NDEBUG

#include <assert.h>

#include "eic.pb.h"
#include "event.h"

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

int main() { serialize1(); }
