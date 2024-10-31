// vi:ft=c

struct Hermite {
    u8 frame_delta;
    u8 control_point;
    u8 control_tangent_0;
    u8 control_tangent_1;
};

struct Keyframe<T, auto Size> {
    u16 position;
    u16 position_delta;
    u16 tangent_0;
    u16 tangent_delta_0;
    u16 tangent_1;
    u16 tangent_delta_1;
    T   curves[Size];
};

struct Record {
    u8  target;
    u8  channel;
    u16 curve_total;
    u32 use_global_transform;
    Keyframe<Hermite, curve_total> *keyframes: u32;
};

struct Mot {
    char signature[4];
    u16  frame_total;
    u8   record_total;
    u8   use_ik;
    Record records[record_total];
};

Mot mot_at_0x00 @ 0x00;
