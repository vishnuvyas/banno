const std = @import("std");

fn manageArgs(alloc: std.mem.Allocator) void {
    const args = std.process.argsAlloc(alloc) catch |err| {
        std.debug.print("Error encounterd while processing args {any} \n", .{err});
        return;
    };
    defer std.process.argsFree(alloc, args); 
    for (args) |arg| {
        std.debug.print("Argument {s} \n", .{arg}); 
    }
}

pub fn main() !void {
    std.debug.print("Hello, Zig World", .{});
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    const allocator = gpa.allocator();
    defer _ = gpa.deinit();

    manageArgs(allocator);

}
