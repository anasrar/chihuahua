// vi:ft=c

struct Name {
    char value[8];
};

struct Tm3 {
    char signature[4];
    u32  total;
    padding[8];
    u32  offsets[total];
    if (total & 0x1 == 1) {
        padding[4];
    }
    Name names[total];
};

Tm3 tm3_at_0x00 @ 0x00;
