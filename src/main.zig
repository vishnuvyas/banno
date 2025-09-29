const std = @import("std");

const Tool = enum { todo, mail };

fn manageArgs(alloc: std.mem.Allocator) !void {
    const args = std.process.argsAlloc(alloc) catch |err| {
        std.debug.print("Error encounterd while processing args {any} \n", .{err});
        return;
    };
    defer std.process.argsFree(alloc, args);
    if (args.len > 1) {
        const tool = args[1];
        const toolEnum = std.meta.stringToEnum(Tool, tool) orelse {
            return error.InvalidToolName;
        };
        switch (toolEnum) {
            Tool.todo => {
                std.debug.print("Provided todo tool\n", .{});
                const toolArgs = args[2..];
                for (toolArgs) |toolArg| {
                    std.debug.print("Provided tool arg {s}\n", .{toolArg});
                }
            },
            else => {
                std.debug.print("UnImplemented tool {any}", .{toolEnum});
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
