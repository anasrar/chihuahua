// vi:ft=c

struct Color {
    u8 red   [[color("FF0000")]];
    u8 green [[color("00FF00")]];
    u8 blue  [[color("0000FF")]];
    padding[1];
};


struct T32 {
    u8    unknown_0[12]; // NOTE: header image data
    u32   clut_offset; // NOTE: relative to image_data
    u8    unknown_1[208];
    u8    image_data[clut_offset - 256];
    u8    unknown_2[256]; // NOTE: header clut data
    Color clut_data[256];
};

T32 t32_at_0x00 @ 0x00;
