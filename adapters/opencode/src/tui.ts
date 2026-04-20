/**
 * opencode-reflector TUI Plugin.
 * Registers slash commands for manual reflection and cleanup.
 */
import type { TuiPlugin, TuiCommand, TuiPluginApi } from "@opencode-ai/plugin/tui";

const reflectorTuiPlugin: TuiPlugin = async (api: TuiPluginApi) => {
  // Register slash commands
  const unregisterCommands = api.command.register(() => {
    const commands: TuiCommand[] = [
      {
        title: "Reflect Now",
        value: "or:reflect_now",
        description: "立即执行反思分析，分析当前和历史会话数据",
        category: "Reflector",
        slash: {
          name: "or:reflect_now",
          aliases: ["or:reflect", "or:rf"],
        },
        onSelect() {
          // Trigger the reflect tool via opencode's command system
          api.command.trigger("or:reflect_now");
        },
      },
      {
        title: "Cleanup Reflector Data",
        value: "or:cleanup",
        description: "清理旧的反思数据（默认保留90天）",
        category: "Reflector",
        slash: {
          name: "or:cleanup",
          aliases: ["or:clean"],
        },
        onSelect() {
          api.command.trigger("or:cleanup");
        },
      },
    ];
    return commands;
  });

  // Cleanup on dispose
  api.lifecycle.onDispose(() => {
    unregisterCommands();
  });
};

export default reflectorTuiPlugin;
