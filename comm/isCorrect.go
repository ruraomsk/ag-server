package comm

import "github.com/ruraomsk/ag-server/pudge"

func isCorectCommand(command pudge.CommandARM) bool {
	if !(command.Command == 6 || command.Command == 7) {
		return true
	}
	if command.Params == 0 {
		return true
	}
	reg := pudge.GetRegion(command.ID)
	if reg.Region == 0 {
		return false
	}
	cr, is := pudge.GetCross(reg)
	if !is {
		return false
	}
	if command.Command == 6 {
		//Смена суточной карты
		for _, v := range cr.Arrays.DaySets.DaySets {
			if v.Number == command.Params {
				if v.Count != 0 {
					return true
				} else {
					return false
				}
			}
		}
		return false
	}
	if command.Command == 7 {
		//Смена  недельной карты
		for _, v := range cr.Arrays.WeekSets.WeekSets {
			if v.Number == command.Params {
				for _, d := range v.Days {
					if d != 0 {
						return true
					}
				}
				return false
			}
		}
		return false
	}
	return true
}
