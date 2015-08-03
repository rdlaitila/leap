local mu = leap.Mutex()
local wg = leap.WaitGroup()
local ct = 0

for a=1, 100 do
    wg:add(1)
    leap.Thread(function()
        wg:done()
    end):run()
end

wg:wait()

print(ct)