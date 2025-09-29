<script>
  import * as Select from "$lib/components/ui/select";
  import * as Card from "$lib/components/ui/card";
  import { setMode, systemPrefersMode, theme } from "mode-watcher";

  const themes = [
    {
      value: "system",
      label: "System",
    },
    {
      value: "dark",
      label: "Dark",
    },
    {
      value: "light",
      label: "Light",
    },
  ];

  function handleThemeChange(theme) {
    currTheme = theme;
    if (theme === "system") {
      setMode(systemPrefersMode.current);
    } else {
      setMode(theme);
    }
  }

  let currTheme = $state("system");
  const triggerContent = $derived(
    themes.find((t) => t.value === currTheme)?.label ?? "system",
  );
</script>

<div class="p-5">
  <Card.Root class="p-0 gap-2">
    <Card.Header class=" flex flex-row items-center justify-center p-1">
      <div class=" flex flex-row items-center justify-center p-0.5 w-full mx-3">
        <Card.Title class="grow text-left">Theme</Card.Title>
        <Select.Root type="single" onValueChange={handleThemeChange}>
          <Select.Trigger class="w-[180px]">
            {triggerContent}
          </Select.Trigger>
          <Select.Content>
            {#each themes as theme}
              <Select.Item value={theme.value}>{theme.label}</Select.Item>
            {/each}
          </Select.Content>
        </Select.Root>
      </div>
    </Card.Header>
  </Card.Root>
</div>
