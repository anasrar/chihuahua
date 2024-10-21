// vi:ft=c

struct Type {
    char name[4];
};

struct Dat {
    u32  total          [[color("16a34a")]];
    u32  offsets[total] [[color("4f46e5")]]; // NOTE: relative to total
    Type types[total]   [[color("ca8a04")]];
};

Dat dat_at_0x00 @ 0x00;
