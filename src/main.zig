const std = @import("std");
const flags = @import("./flags.zig");
const Tool = enum { todo, mail };

fn todo_tool() !void {
    const flag_struct = flags.flag_int("n", "number");
    std.debug.print("Provided following flag sturct : {any}", .{flag_struct});
}

fn manageArgs(alloc: std.mem.Allocator) !void {
    const args = std.process.argsAlloc(alloc) catch |err| {
        std.debug.print("Error encounterd while processing args {any} \n", .{err});
        return;
    };
    defer std.process.argsFree(alloc, args);
    if (args.len > 1) {
        const tool = args[1];
        const tool_enum = std.meta.stringToEnum(Tool, tool) orelse {
            return error.InvalidToolName;
        };
        switch (tool_enum) {
            Tool.todo => {
                std.debug.print("Provided todo tool\n", .{});
                try todo_tool();
            },
            else => {
                std.debug.print("UnImplemented tool {any}", .{tool_enum});
            },
        }
    }
}

pub fn main() !void {
    std.debug.print("Hello, Zig World\n", .{});
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    const allocator = gpa.allocator();
    defer _ = gpa.deinit();

    try manageArgs(allocator);
}
