#include <iostream>

#include <TFile.h>
#include <TTree.h>

#include "eicio/reader.h"

using namespace google::protobuf;

int main(int argc, char **argv) {
    if (argc < 3) return EXIT_FAILURE;  // TODO: implement getopts

    eicio::Reader *reader = new eicio::Reader(argv[1]);
    TFile oFile(argv[2], "recreate");
    std::vector<TTree *> trees;
    std::map<std::string, std::map<std::string, void *>> fieldVars;

    while (eicio::Event *event = reader->Get()) {
        for (auto name : event->GetNames()) {
            Message *coll = event->Get(name);
            if (coll != NULL) {
                const Descriptor *desc = coll->GetDescriptor();
                const Reflection *ref = coll->GetReflection();

                const RepeatedPtrField<Message> entries = ref->GetRepeatedPtrField<Message>(
                    *((const Message *)coll), desc->FindFieldByName("entries"));
                for (int i = 0; i < entries.size(); i++) {
                    std::vector<const FieldDescriptor *> fieldDescs;
                    const Reflection *entryRef = entries[i].GetReflection();
                    entryRef->ListFields(entries[i], &fieldDescs);

                    for (auto fieldDesc : fieldDescs) {
                        if (fieldVars[name].find(fieldDesc->name()) == fieldVars[name].end()) {
                            switch (fieldDesc->type()) {
                                case FieldDescriptor::TYPE_UINT32:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new uint32;
                                    }
                                    break;
                                case FieldDescriptor::TYPE_UINT64:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new uint64;
                                    }
                                    break;
                                case FieldDescriptor::TYPE_FLOAT:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new float;
                                    }
                                    break;
                                case FieldDescriptor::TYPE_DOUBLE:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new double;
                                    }
                                    break;
                                case FieldDescriptor::TYPE_INT32:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new int32;
                                    }
                                    break;
                                case FieldDescriptor::TYPE_INT64:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new int64;
                                    }
                                    break;
                            }
                        }

                        switch (fieldDesc->type()) {
                            case FieldDescriptor::TYPE_UINT32:
                                if (!fieldDesc->is_repeated()) {
                                    *((uint32 *)fieldVars[name][fieldDesc->name()]) =
                                        entryRef->GetUInt32(entries[i], fieldDesc);
                                }
                                break;
                            case FieldDescriptor::TYPE_UINT64:
                                if (!fieldDesc->is_repeated()) {
                                    *((uint64 *)fieldVars[name][fieldDesc->name()]) =
                                        entryRef->GetUInt64(entries[i], fieldDesc);
                                }
                                break;
                            case FieldDescriptor::TYPE_FLOAT:
                                if (!fieldDesc->is_repeated()) {
                                    *((float *)fieldVars[name][fieldDesc->name()]) =
                                        entryRef->GetFloat(entries[i], fieldDesc);
                                }
                                break;
                            case FieldDescriptor::TYPE_DOUBLE:
                                if (!fieldDesc->is_repeated()) {
                                    *((double *)fieldVars[name][fieldDesc->name()]) =
                                        entryRef->GetDouble(entries[i], fieldDesc);
                                }
                                break;
                            case FieldDescriptor::TYPE_INT32:
                                if (!fieldDesc->is_repeated()) {
                                    *((int32 *)fieldVars[name][fieldDesc->name()]) =
                                        entryRef->GetInt32(entries[i], fieldDesc);
                                }
                                break;
                            case FieldDescriptor::TYPE_INT64:
                                if (!fieldDesc->is_repeated()) {
                                    *((int64 *)fieldVars[name][fieldDesc->name()]) =
                                        entryRef->GetInt64(entries[i], fieldDesc);
                                }
                                break;
                        }
                    }

                    auto tree = (TTree *)oFile.Get(name.c_str());
                    if (tree == NULL) {
                        tree = new TTree(name.c_str(), desc->full_name().c_str());
                        trees.push_back(tree);

                        for (auto fieldDesc : fieldDescs) {
                            switch (fieldDesc->type()) {
                                case FieldDescriptor::TYPE_UINT32:
                                    if (!fieldDesc->is_repeated()) {
                                        tree->Branch(fieldDesc->name().c_str(),
                                                     fieldVars[name][fieldDesc->name()], "uint32/i");
                                    }
                                    break;
                                case FieldDescriptor::TYPE_UINT64:
                                    if (!fieldDesc->is_repeated()) {
                                        tree->Branch(fieldDesc->name().c_str(),
                                                     fieldVars[name][fieldDesc->name()], "uint64/l");
                                    }
                                    break;
                                case FieldDescriptor::TYPE_FLOAT:
                                    if (!fieldDesc->is_repeated()) {
                                        tree->Branch(fieldDesc->name().c_str(),
                                                     fieldVars[name][fieldDesc->name()], "float/F");
                                    }
                                    break;
                                case FieldDescriptor::TYPE_DOUBLE:
                                    if (!fieldDesc->is_repeated()) {
                                        tree->Branch(fieldDesc->name().c_str(),
                                                     fieldVars[name][fieldDesc->name()], "double/D");
                                    }
                                    break;
                                case FieldDescriptor::TYPE_INT32:
                                    if (!fieldDesc->is_repeated()) {
                                        tree->Branch(fieldDesc->name().c_str(),
                                                     fieldVars[name][fieldDesc->name()], "int32/I");
                                    }
                                    break;
                                case FieldDescriptor::TYPE_INT64:
                                    if (!fieldDesc->is_repeated()) {
                                        tree->Branch(fieldDesc->name().c_str(),
                                                     fieldVars[name][fieldDesc->name()], "int64/L");
                                    }
                                    break;
                            }
                        }
                    }

                    tree->Fill();
                }
            }
        }

        delete event;
    }

    for (auto tree : trees) {
        tree->Write();
    }
    oFile.Close();
    delete reader;
    return EXIT_SUCCESS;
}
