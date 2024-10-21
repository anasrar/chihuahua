// vi:ft=c

struct Color {
    u8 red   [[color("FF0000")]];
    u8 green [[color("00FF00")]];
    u8 blue  [[color("0000FF")]];
    padding[1];
};

struct FileHeader {
    char signature[4];
    u8   format_version;
    u8   format_id;
    u16  pictures;
    u8   reserved[8];
};

struct MipMap {
    u64 gs_miptbp1;
    u64 gs_miptbp2;
    u32 size[1];
};

struct Picture {
    u32 total_size;
    u32 clut_size;
    u32 image_size;
    u16 header_size;
    u16 clut_colors;
    u8  pict_format;
    u8  mipmap_textures;
    u8  clut_type;
    u8  image_type;
    u16 image_width;
    u16 image_height;
    u64 gs_tex0;
    u64 gs_tex1;
    u32 gs_regs;
    u32 gs_tex_clut;

    // NOTE: seem god hand does not have mipmap
    // MipMap mipmaps; // NOTE: change size

    u8    image_data[image_size];
    Color clut_data[clut_size/4];
};

struct Tim2 {
    FileHeader header;
    Picture    picture;
};

Tim2 tim2_at_0x00 @ 0x00;
