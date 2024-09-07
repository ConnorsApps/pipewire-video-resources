-- ~/.config/wireplumber/scripts/auto-setup-ports.lua
--
-- Main Source
-- https://bennett.dev/auto-link-pipewire-ports-wireplumber/
-- https://github.com/bennetthardwick/dotfiles/blob/master/.config/wireplumber/scripts/auto-connect-ports.lua
--
-- Pipewire Docs
-- https://docs.pipewire.org/index.html


local function link_port(output_port, input_port)
	if not input_port or not output_port then
	  return nil
	end
 
	local link_args = {
	  ["link.input.node"] = input_port.properties["node.id"],
	  ["link.input.port"] = input_port.properties["object.id"],
 
	  ["link.output.node"] = output_port.properties["node.id"],
	  ["link.output.port"] = output_port.properties["object.id"],
	  
	  -- The node never got created if it didn't have this field set to something
	  ["object.id"] = nil,
 
	  -- I was running into issues when I didn't have this set
	  ["object.linger"] = true,
 
	  ["node.description"] = "Link created by auto_connect_ports"
	}
 
	local link = Link("link-factory", link_args)
	link:activate(1)

	return link
 end

-- Left in for later debugging
local function dump(o)
   if type(o) == 'table' then
      local s = '{ '
      for k,v in pairs(o) do
         if type(k) ~= 'number' then k = '"'..k..'"' end
         s = s .. '['..k..'] = ' .. dump(v) .. ',\n'
      end
      return s .. '} '
   else
      return tostring(o)
   end
end

local function get_input(constraints)
	local interest = {
			type = "port",
			Constraint { "port.direction", "equals", "in" },
	}

	for _, constraint in ipairs(constraints) do
		table.insert(interest, constraint)
	end

	return ObjectManager {
		Interest(interest)
	}
end

local function get_output(constraints)
	local interest = {
			type = "port",
			Constraint { "port.direction", "equals", "out" },
	}

	for _, constraint in ipairs(constraints) do
		table.insert(interest, constraint)
	end

	return ObjectManager {
		Interest(interest)
	}
end

local function auto_connect_ports(args)
	local input = args["input"]
	local output = args["output"]

	local all_links = ObjectManager {
		Interest {
		  type = "link",
		}
	 }
	
	
	local function lookup_link(output_port, input_port)
		return all_links:lookup {
			Constraint { "link.output.port", "=", output_port.properties["object.id"]},
			Constraint { "link.input.port", "=", input_port.properties["object.id"]},
			Constraint { "link.output.node", "=", output_port.properties["node.id"]},
			Constraint { "link.input.node", "=", input_port.properties["node.id"]},
		}
	end
	
	-- Each time there is an event on any port
	-- try to re-establish the connection
	local function _object_added()
		local channels = {"FL", "FR"}

		for _, channel in pairs(channels) do
			local output_port = output:lookup {
				Constraint { "audio.channel", "equals", channel },
			}
			local input_port =  input:lookup {
				Constraint { "audio.channel", "equals", channel },
			}

			if output_port == nil or input_port == nil then
				goto continue
			end

			-- print("output_port",  dump(output_port.properties))
			-- print("input_port", dump(input_port.properties))

			-- Check if link already exists
			if lookup_link(output_port, input_port) ~= nil then
				print("Link already exists for", output_port.properties["port.alias"], "->",input_port.properties["port.alias"])
				goto continue
			end
	
			print("Linking", output_port.properties["port.alias"], "->",input_port.properties["port.alias"])

			link_port(output_port, input_port)
		    ::continue::
		end
	end

	input:connect("object-added", function (om, port)
		-- print("Input Added", dump(port.properties["port.alias"]))
		_object_added()
	end)

	output:connect("object-added", function (om, port)
		-- print("Output Added", dump(port.properties["port.alias"]))
		_object_added()
	end)

	-- all_links:connect("object-added", function (om, port)
	-- 	_object_added()
	-- end)

	input:activate()
	output:activate()
	all_links:activate()
end


volt = get_input {
	Constraint { "format.dsp", "not-equals", "8 bit raw midi" },
	Constraint { "port.alias", "matches", "Volt 276:*" },
	Constraint { "port.monitor", "not-equals", "true" }
}

record_player = get_output {
	Constraint { "port.alias", "matches", "USB AUDIO  CODEC:*" },
	Constraint { "port.monitor", "not-equals", "true" }
}

spotifyd = get_output {
	Constraint { "port.alias", "matches", "ALSA Playback [spotifyd]:*" },
}


auto_connect_ports {
	output = record_player,
	input = volt,
}

auto_connect_ports {
	output = spotifyd,
	input = volt,
}
