-- Lumina Example: Counter (Updated)
-- Showcases: defineComponent, init/render, styling, theme colors, progress bar
local lumina = require("lumina")
local theme = {
    bg="#1E1E2E", fg="#CDD6F4", accent="#89B4FA", success="#A6E3A1",
    error="#F38BA8", muted="#6C7086", headerBg="#181825",
}
local Counter = lumina.defineComponent({
    name = "Counter",
    init = function(props)
        return { count = props.initial or 0, step = props.step or 1,
                 min = props.min or 0, max = props.max or 999 }
    end,
    render = function(inst)
        local count = inst.count
        local valueColor = theme.fg
        if count > 50 then valueColor = theme.error
        elseif count > 20 then valueColor = "#F9E2AF"
        elseif count > 0 then valueColor = theme.success end
        local barW = 30
        local range = inst.max - inst.min
        local filled = range > 0 and math.floor(((count - inst.min) / range) * barW + 0.5) or 0
        filled = math.max(0, math.min(barW, filled))
        local bar = string.rep("█", filled) .. string.rep("░", barW - filled)
        return {type="vbox",style={flex=1,background=theme.bg,padding=1},children={
            {type="text",content="╔══════════════════════════════════╗",style={foreground=theme.accent}},
            {type="text",content="║       Lumina Counter             ║",style={foreground=theme.accent,bold=true}},
            {type="text",content="╚══════════════════════════════════╝",style={foreground=theme.accent}},
            {type="text",content=""},
            {type="hbox",children={
                {type="text",content="  Value: ",style={foreground=theme.muted}},
                {type="text",content=tostring(count),style={foreground=valueColor,bold=true}},
                {type="text",content=string.format("  (step: %d)",inst.step),style={foreground=theme.muted}},
            }},
            {type="text",content=""},
            {type="hbox",children={
                {type="text",content="  ["..bar.."]",style={foreground=valueColor}},
                {type="text",content=string.format(" %d/%d",count,inst.max),style={foreground=theme.muted}},
            }},
            {type="text",content=""},
            {type="hbox",style={gap=2},children={
                {type="text",content="  [ - ] Decrement",style={foreground=theme.error}},
                {type="text",content="  [ + ] Increment",style={foreground=theme.success}},
                {type="text",content="  [ r ] Reset",style={foreground=theme.muted}},
            }},
            {type="text",content=""},
            {type="hbox",style={height=1,background=theme.headerBg},children={
                {type="text",content=string.format(" Range: %d-%d  |  Step: %d ",inst.min,inst.max,inst.step),style={foreground=theme.muted}},
                {type="text",content=" [q] Quit ",style={foreground=theme.muted}},
            }},
        }}
    end,
})
lumina.render(Counter, { initial = 42, step = 5, max = 100 })
