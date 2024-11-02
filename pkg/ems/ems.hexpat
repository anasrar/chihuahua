struct Entry {
    u32   unknown_0     [[color("800000")]];
    float translation_x [[color("FF0000")]];
    float translation_y [[color("00FF00")]];
    float translation_z [[color("0000FF")]];
    u8    unknown_1[48] [[color("303030")]];
};

struct Ems {
    char  signature[4];
    u32   entry_total;
    Entry entries[entry_total];
};

Ems ems_at_0x00 @ 0x00;
