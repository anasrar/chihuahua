// vi:ft=c

struct Node {
  s32   mdb_offset; // NOTE: relative to this Node position
  padding[4]; // TODO: research this padding
  char  name[8];
  float scale_x;
  float scale_y;
  float scale_z;
  float rotation_x;
  float rotation_y;
  float rotation_z;
  float translation_x;
  float translation_y;
  float translation_z;
  padding[60]; // TODO: research this padding
};

struct Scr {
  char signature[4];
  padding[4]; // TODO: research this padding
  u32  node_total;
  padding[4]; // TODO: research this padding
  Node *node_offsets[node_total] : s32; // NOTE: relative to Scr position
};

Scr scr_at_0x00 @ 0x00;
