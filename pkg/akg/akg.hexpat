struct Entry0 {
    u16  index      [[color("FF0000")]];
    u16  unknown[5] [[color("808080")]];
    char name[8]    [[color("00FF00")]];
};

struct Entry1 {
    u16 unknown[8] [[color("808080")]];
};

struct Entry7 {
    char name[8] [[color("00FF00")]];
};

struct Akg {
    u32 offset_0;
    u32 offset_1;
    u32 offset_2;
    u32 offset_3;
    u32 offset_4;
    u32 offset_5;
    u32 offset_6;
    u32 offset_7;
    u16 entry_0_total;
    u16 entry_1_total;
    u16 entry_2_total;
    u16 entry_3_total;
    u16 entry_4_total;
    u16 entry_5_total;
    Entry0 entries_0[entry_0_total] @ offset_0;
    Entry1 entries_1[entry_1_total] @ offset_1;
    Entry7 entries_7[entry_5_total] @ offset_7;
};

Akg akg_at_0x00 @ 0x00;
