#include <iostream>

#include <TFile.h>
#include <TTree.h>

#include "eicio/reader.h"

int main(int argc, char **argv) {
    if (argc < 3) return EXIT_FAILURE;  // TODO: implement getopts

    eicio::Reader *reader = new eicio::Reader(argv[1]);
    TFile oFile(argv[2], "recreate");
    std::vector<TTree *> trees;
    std::map<std::string, std::map<std::string, void *>> fieldVars;

    while (eicio::Event *event = reader->Get()) {
        for (auto name : event->GetNames()) {
            google::protobuf::Message *coll = event->Get(name);
            if (coll != NULL) {
                const google::protobuf::Descriptor *desc = coll->GetDescriptor();
                const google::protobuf::Reflection *ref = coll->GetReflection();

                const google::protobuf::RepeatedPtrField<google::protobuf::Message> entries =
                    ref->GetRepeatedPtrField<google::protobuf::Message>(
                        *((const google::protobuf::Message *)coll), desc->FindFieldByName("entries"));
                for (int i = 0; i < entries.size(); i++) {
                    std::vector<const google::protobuf::FieldDescriptor *> fieldDescs;
                    const google::protobuf::Reflection *entryRef = entries[i].GetReflection();
                    entryRef->ListFields(entries[i], &fieldDescs);

                    for (auto fieldDesc : fieldDescs) {
                        if (fieldVars[name].find(fieldDesc->name()) == fieldVars[name].end()) {
                            switch (fieldDesc->type()) {
                                case google::protobuf::FieldDescriptor::TYPE_UINT32:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new google::protobuf::uint32;
                                    }
                                    break;
                                case google::protobuf::FieldDescriptor::TYPE_UINT64:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new google::protobuf::uint64;
                                    }
                                    break;
                                case google::protobuf::FieldDescriptor::TYPE_FLOAT:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new float;
                                    }
                                    break;
                                case google::protobuf::FieldDescriptor::TYPE_DOUBLE:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new double;
                                    }
                                    break;
                                case google::protobuf::FieldDescriptor::TYPE_INT32:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new google::protobuf::int32;
                                    }
                                    break;
                                case google::protobuf::FieldDescriptor::TYPE_INT64:
                                    if (!fieldDesc->is_repeated()) {
                                        fieldVars[name][fieldDesc->name()] = new google::protobuf::int64;
                                    }
                                    break;
                            }
                        }

                        switch (fieldDesc->type()) {
                            case google::protobuf::FieldDescriptor::TYPE_UINT32:
                                if (!fieldDesc->is_repeated()) {
                                    *((google::protobuf::uint32 *)fieldVars[name][fieldDesc->name()]) =
                                        entryRef->GetUInt32(entries[i], fieldDesc);
                                }
                                break;
                            case google::protobuf::FieldDescriptor::TYPE_UINT64:
                                if (!fieldDesc->is_repeated()) {
                                    *((google::protobuf::uint64 *)fieldVars[name][fieldDesc->name()]) =
                                        entryRef->GetUInt64(entries[i], fieldDesc);
                                }
                                break;
                            case google::protobuf::FieldDescriptor::TYPE_FLOAT:
                                if (!fieldDesc->is_repeated()) {
                                    *((float *)fieldVars[name][fieldDesc->name()]) =
                                        entryRef->GetFloat(entries[i], fieldDesc);
                                }
                                break;
                            case google::protobuf::FieldDescriptor::TYPE_DOUBLE:
                                if (!fieldDesc->is_repeated()) {
                                    *((double *)fieldVars[name][fieldDesc->name()]) =
                                        entryRef->GetDouble(entries[i], fieldDesc);
                                }
                                break;
                            case google::protobuf::FieldDescriptor::TYPE_INT32:
                                if (!fieldDesc->is_repeated()) {
                                    *((google::protobuf::int32 *)fieldVars[name][fieldDesc->name()]) =
                                        entryRef->GetInt32(entries[i], fieldDesc);
                                }
                                break;
                            case google::protobuf::FieldDescriptor::TYPE_INT64:
                                if (!fieldDesc->is_repeated()) {
                                    *((google::protobuf::int64 *)fieldVars[name][fieldDesc->name()]) =
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
                                case google::protobuf::FieldDescriptor::TYPE_UINT32:
                                    if (!fieldDesc->is_repeated()) {
                                        tree->Branch(fieldDesc->name().c_str(),
                                                     fieldVars[name][fieldDesc->name()], "uint32/i");
                                    }
                                    break;
                                case google::protobuf::FieldDescriptor::TYPE_UINT64:
                                    if (!fieldDesc->is_repeated()) {
                                        tree->Branch(fieldDesc->name().c_str(),
                                                     fieldVars[name][fieldDesc->name()], "uint64/l");
                                    }
                                    break;
                                case google::protobuf::FieldDescriptor::TYPE_FLOAT:
                                    if (!fieldDesc->is_repeated()) {
                                        tree->Branch(fieldDesc->name().c_str(),
                                                     fieldVars[name][fieldDesc->name()], "float/F");
                                    }
                                    break;
                                case google::protobuf::FieldDescriptor::TYPE_DOUBLE:
                                    if (!fieldDesc->is_repeated()) {
                                        tree->Branch(fieldDesc->name().c_str(),
                                                     fieldVars[name][fieldDesc->name()], "double/D");
                                    }
                                    break;
                                case google::protobuf::FieldDescriptor::TYPE_INT32:
                                    if (!fieldDesc->is_repeated()) {
                                        tree->Branch(fieldDesc->name().c_str(),
                                                     fieldVars[name][fieldDesc->name()], "int32/I");
                                    }
                                    break;
                                case google::protobuf::FieldDescriptor::TYPE_INT64:
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
