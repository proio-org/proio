#include <unistd.h>
#include <iostream>

#include <TFile.h>
#include <TTree.h>

#include "eicio/reader.h"

using namespace google::protobuf;

void printUsage() {
    std::cerr << "Usage: eicio2root [options] <input eicio file> <output root file>\n";
    std::cerr << "options:\n";
    std::cerr << "  -g	decompress the stdin input with gzip\n";
    std::cerr << std::endl;
}

int main(int argc, char **argv) {
    int opt;
    bool gzip = false;
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

    TFile oFile(outputFilename.c_str(), "recreate");
    std::vector<TTree *> trees;
    std::map<std::string, std::map<std::string, void *>> fieldVars;

    TTree eventTree("events", "Events");
    std::vector<std::string> containerNames;
    std::vector<unsigned int> containerEntries;
    eventTree.Branch("containerNames", "std::vector<std::string>", &containerNames);
    eventTree.Branch("containerEntries", "std::vector<unsigned int>", &containerEntries);

    while (eicio::Event *event = reader->Get()) {
        containerNames.clear();
        containerEntries.clear();

        for (auto name : event->GetNames()) {
            Message *coll = event->Get(name);
            if (!coll) continue;

            const Descriptor *desc = coll->GetDescriptor();
            const Reflection *ref = coll->GetReflection();

            const FieldDescriptor *entriesFieldDesc = desc->FindFieldByName("entries");
            if (!entriesFieldDesc) continue;
            const RepeatedPtrField<Message> entries =
                ref->GetRepeatedPtrField<Message>(*coll, entriesFieldDesc);

            containerNames.push_back(name);
            containerEntries.push_back(entries.size());

            for (int i = 0; i < entries.size(); i++) {
                const Descriptor *entryDesc = entries[i].GetDescriptor();
                const Reflection *entryRef = entries[i].GetReflection();

                for (int j = 0; j < entryDesc->field_count(); j++) {
                    const FieldDescriptor *fieldDesc = entryDesc->field(j);

                    if (fieldVars[name].find(fieldDesc->name()) == fieldVars[name].end()) {
                        switch (fieldDesc->type()) {
                            case FieldDescriptor::TYPE_UINT32:
                                fieldVars[name][fieldDesc->name()] = fieldDesc->is_repeated()
                                                                         ? (void *)new std::vector<uint32>
                                                                         : (void *)new uint32;
                                break;
                            case FieldDescriptor::TYPE_UINT64:
                                fieldVars[name][fieldDesc->name()] = fieldDesc->is_repeated()
                                                                         ? (void *)new std::vector<uint64>
                                                                         : (void *)new uint64;
                                break;
                            case FieldDescriptor::TYPE_INT32:
                                fieldVars[name][fieldDesc->name()] = fieldDesc->is_repeated()
                                                                         ? (void *)new std::vector<int32>
                                                                         : (void *)new int32;
                                break;
                            case FieldDescriptor::TYPE_INT64:
                                fieldVars[name][fieldDesc->name()] = fieldDesc->is_repeated()
                                                                         ? (void *)new std::vector<int64>
                                                                         : (void *)new int64;
                                break;
                            case FieldDescriptor::TYPE_FLOAT:
                                fieldVars[name][fieldDesc->name()] = fieldDesc->is_repeated()
                                                                         ? (void *)new std::vector<float>
                                                                         : (void *)new float;
                                break;
                            case FieldDescriptor::TYPE_DOUBLE:
                                fieldVars[name][fieldDesc->name()] = fieldDesc->is_repeated()
                                                                         ? (void *)new std::vector<double>
                                                                         : (void *)new double;
                                break;
                        }
                    }

                    switch (fieldDesc->type()) {
                        case FieldDescriptor::TYPE_UINT32:
                            if (!fieldDesc->is_repeated()) {
                                *((uint32 *)fieldVars[name][fieldDesc->name()]) =
                                    entryRef->GetUInt32(entries[i], fieldDesc);
                            } else {
                                auto vec = (std::vector<uint32> *)fieldVars[name][fieldDesc->name()];
                                vec->clear();
                                for (int k = 0; k < entryRef->FieldSize(entries[i], fieldDesc); k++)
                                    vec->push_back(entryRef->GetRepeatedUInt32(entries[i], fieldDesc, k));
                            }
                            break;
                        case FieldDescriptor::TYPE_UINT64:
                            if (!fieldDesc->is_repeated()) {
                                *((uint64 *)fieldVars[name][fieldDesc->name()]) =
                                    entryRef->GetUInt64(entries[i], fieldDesc);
                            } else {
                                auto vec = (std::vector<uint64> *)fieldVars[name][fieldDesc->name()];
                                vec->clear();
                                for (int k = 0; k < entryRef->FieldSize(entries[i], fieldDesc); k++)
                                    vec->push_back(entryRef->GetRepeatedUInt64(entries[i], fieldDesc, k));
                            }
                            break;
                        case FieldDescriptor::TYPE_INT32:
                            if (!fieldDesc->is_repeated()) {
                                *((int32 *)fieldVars[name][fieldDesc->name()]) =
                                    entryRef->GetInt32(entries[i], fieldDesc);
                            } else {
                                auto vec = (std::vector<int32> *)fieldVars[name][fieldDesc->name()];
                                vec->clear();
                                for (int k = 0; k < entryRef->FieldSize(entries[i], fieldDesc); k++)
                                    vec->push_back(entryRef->GetRepeatedInt32(entries[i], fieldDesc, k));
                            }
                            break;
                        case FieldDescriptor::TYPE_INT64:
                            if (!fieldDesc->is_repeated()) {
                                *((int64 *)fieldVars[name][fieldDesc->name()]) =
                                    entryRef->GetInt64(entries[i], fieldDesc);
                            } else {
                                auto vec = (std::vector<int64> *)fieldVars[name][fieldDesc->name()];
                                vec->clear();
                                for (int k = 0; k < entryRef->FieldSize(entries[i], fieldDesc); k++)
                                    vec->push_back(entryRef->GetRepeatedInt64(entries[i], fieldDesc, k));
                            }
                            break;
                        case FieldDescriptor::TYPE_FLOAT:
                            if (!fieldDesc->is_repeated()) {
                                *((float *)fieldVars[name][fieldDesc->name()]) =
                                    entryRef->GetFloat(entries[i], fieldDesc);
                            } else {
                                auto vec = (std::vector<float> *)fieldVars[name][fieldDesc->name()];
                                vec->clear();
                                for (int k = 0; k < entryRef->FieldSize(entries[i], fieldDesc); k++)
                                    vec->push_back(entryRef->GetRepeatedFloat(entries[i], fieldDesc, k));
                            }
                            break;
                        case FieldDescriptor::TYPE_DOUBLE:
                            if (!fieldDesc->is_repeated()) {
                                *((double *)fieldVars[name][fieldDesc->name()]) =
                                    entryRef->GetDouble(entries[i], fieldDesc);
                            } else {
                                auto vec = (std::vector<double> *)fieldVars[name][fieldDesc->name()];
                                vec->clear();
                                for (int k = 0; k < entryRef->FieldSize(entries[i], fieldDesc); k++)
                                    vec->push_back(entryRef->GetRepeatedDouble(entries[i], fieldDesc, k));
                            }
                            break;
                    }
                }

                auto tree = (TTree *)oFile.Get(name.c_str());
                if (tree == NULL) {
                    tree = new TTree(name.c_str(), desc->full_name().c_str());
                    trees.push_back(tree);

                    for (int j = 0; j < entryDesc->field_count(); j++) {
                        const FieldDescriptor *fieldDesc = entryDesc->field(j);

                        switch (fieldDesc->type()) {
                            case FieldDescriptor::TYPE_UINT32:
                                if (!fieldDesc->is_repeated())
                                    tree->Branch(fieldDesc->name().c_str(),
                                                 fieldVars[name][fieldDesc->name()], "uint32/i");
                                else
                                    tree->Branch(fieldDesc->name().c_str(), "std::vector<unsigned int>",
                                                 (std::vector<uint32> *)fieldVars[name][fieldDesc->name()]);
                                break;
                            case FieldDescriptor::TYPE_UINT64:
                                if (!fieldDesc->is_repeated())
                                    tree->Branch(fieldDesc->name().c_str(),
                                                 fieldVars[name][fieldDesc->name()], "uint64/l");
                                else
                                    tree->Branch(fieldDesc->name().c_str(), "std::vector<unsigned long>",
                                                 (std::vector<uint64> *)fieldVars[name][fieldDesc->name()]);
                                break;
                            case FieldDescriptor::TYPE_INT32:
                                if (!fieldDesc->is_repeated())
                                    tree->Branch(fieldDesc->name().c_str(),
                                                 fieldVars[name][fieldDesc->name()], "int32/I");
                                else
                                    tree->Branch(fieldDesc->name().c_str(), "std::vector<int>",
                                                 (std::vector<int> *)fieldVars[name][fieldDesc->name()]);
                                break;
                            case FieldDescriptor::TYPE_INT64:
                                if (!fieldDesc->is_repeated())
                                    tree->Branch(fieldDesc->name().c_str(),
                                                 fieldVars[name][fieldDesc->name()], "int64/L");
                                else
                                    tree->Branch(fieldDesc->name().c_str(), "std::vector<long>",
                                                 (std::vector<long> *)fieldVars[name][fieldDesc->name()]);
                                break;
                            case FieldDescriptor::TYPE_FLOAT:
                                if (!fieldDesc->is_repeated())
                                    tree->Branch(fieldDesc->name().c_str(),
                                                 fieldVars[name][fieldDesc->name()], "float/F");
                                else
                                    tree->Branch(fieldDesc->name().c_str(), "std::vector<float>",
                                                 (std::vector<float> *)fieldVars[name][fieldDesc->name()]);
                                break;
                            case FieldDescriptor::TYPE_DOUBLE:
                                if (!fieldDesc->is_repeated())
                                    tree->Branch(fieldDesc->name().c_str(),
                                                 fieldVars[name][fieldDesc->name()], "double/D");
                                else
                                    tree->Branch(fieldDesc->name().c_str(), "std::vector<double>",
                                                 (std::vector<double> *)fieldVars[name][fieldDesc->name()]);
                                break;
                        }
                    }
                }

                tree->Fill();
            }
        }

        delete event;
        eventTree.Fill();
    }

    eventTree.Write();

    for (auto tree : trees) {
        tree->Write();
    }
    oFile.Close();
    delete reader;
    return EXIT_SUCCESS;
}
