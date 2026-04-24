-- Lumina Example: Form Demo (Updated)
-- Showcases: TextInput, Select, Checkbox, form layout, validation hints, theming
local lumina = require("lumina")
local theme = {
    bg="#1E1E2E", fg="#CDD6F4", accent="#89B4FA", success="#A6E3A1",
    error="#F38BA8", muted="#6C7086", surface="#313244", headerBg="#181825", label="#B4BEFE",
}
local function formField(label, child, hint)
    local ch = {
        {type="text",content="  "..label,style={foreground=theme.label,bold=true}},
        child,
    }
    if hint then ch[#ch+1] = {type="text",content="    "..hint,style={foreground=theme.muted}} end
    ch[#ch+1] = {type="text",content=""}
    return {type="vbox",children=ch}
end
local FormDemo = lumina.defineComponent({
    name = "FormDemo",
    init = function(props) return {
        username = "alice", email = "alice@example.com", password = "",
        country = "USA", agreed = false, newsletter = true, role = "developer",
    } end,
    render = function(inst)
        return {type="vbox",style={flex=1,background=theme.bg},children={
            {type="hbox",style={height=1,background=theme.headerBg},children={
                {type="text",content=" Registration Form ",style={foreground=theme.accent,bold=true}},
            }},
            {type="vbox",style={flex=1,overflow="scroll",padding=1,border="rounded",background=theme.bg},children={
                {type="text",content="  -- Account Information --",style={foreground=theme.accent,bold=true}},
                {type="text",content=""},
                formField("Username", {type="hbox",style={height=1,background=theme.surface},children={
                    {type="text",content="  @",style={foreground=theme.muted}},
                    {type="input",id="username",value=inst.username,placeholder="Enter username",style={flex=1,foreground=theme.fg}},
                }}, "3-20 characters, letters and numbers only"),
                formField("Email", {type="hbox",style={height=1,background=theme.surface},children={
                    {type="text",content="  > ",style={foreground=theme.muted}},
                    {type="input",id="email",value=inst.email,placeholder="user@example.com",style={flex=1,foreground=theme.fg}},
                }}, "We'll never share your email"),
                formField("Password", {type="hbox",style={height=1,background=theme.surface},children={
                    {type="text",content="  * ",style={foreground=theme.muted}},
                    {type="input",id="password",value=string.rep("*",#inst.password),placeholder="Enter password",style={flex=1,foreground=theme.fg}},
                }}, "Minimum 8 characters"),
                {type="text",content="  -- Preferences --",style={foreground=theme.accent,bold=true}},
                {type="text",content=""},
                formField("Country",
                    lumina.Select({id="country-select",placeholder="Select country",
                        options={"USA","Canada","UK","Germany","Japan","Australia"},value=inst.country})),
                formField("Role",
                    lumina.Select({id="role-select",placeholder="Select role",
                        options={"developer","designer","manager","other"},value=inst.role})),
                formField("Agreement",
                    lumina.Checkbox({id="agree-checkbox",label="I agree to the Terms of Service",checked=inst.agreed})),
                formField("Newsletter",
                    lumina.Checkbox({id="newsletter-checkbox",label="Subscribe to newsletter",checked=inst.newsletter})),
                {type="text",content=""},
                {type="hbox",style={gap=2,padding=1},children={
                    {type="text",content=" [ Submit ] ",style={foreground=theme.success,bold=true}},
                    {type="text",content=" [ Reset ] ",style={foreground=theme.muted}},
                    {type="text",content=" [ Cancel ] ",style={foreground=theme.error}},
                }},
            }},
            {type="hbox",style={height=1,background=theme.headerBg},children={
                {type="text",content=" [Tab] Next field  [Shift+Tab] Prev  [Enter] Submit  [Esc] Cancel ",style={foreground=theme.muted}},
            }},
        }}
    end,
})
lumina.render(FormDemo, {})
