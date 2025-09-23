const std = @import("std");

fn manageArgs(alloc: std.mem.Allocator) void {
    const args = std.process.argsAlloc(alloc);
}

pub fn main() !void {
    std.debug.print("Hello, Zig World", .{});
}
