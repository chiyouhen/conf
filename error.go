// Copyright 2015, 2016 ZHANG Heng (chiyouhen@gmail.com)
//
//  This file is part of Conf.
//  
//  Conf is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  any later version.
//  
//  Conf is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//  
//  You should have received a copy of the GNU General Public License
//  along with Conf.  If not, see <http://www.gnu.org/licenses/>.

package conf

import (
    "fmt"
)

type ConfError struct {
    message string
    filename string
    lineno int
}

func (e *ConfError) Error() string {
    return fmt.Sprintf("%s in '%s' line %d", e.message, e.filename, e.lineno)
}
