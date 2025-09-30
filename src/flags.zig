const std = @import("std");

const FlagStruct = struct {
    flag_name: []const u8,
    flag_long_name: []const u8,
};

pub fn flag_int(short_name: []const u8, long_name: []const u8) FlagStruct {
    return FlagStruct{
        .flag_name = short_name,
        .flag_long_name = long_name,
    };
}
