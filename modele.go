package main

import (
	"fmt"
	"strings"
)

type Modele struct {
	id    string
	desm  map[int][]*Des
	lgenR []Genrad
	pere  *Modele
	abs   []int
	suf   []string
	pos   string
	sufd  string
}

func (m Modele) doc() string {
	var mid string
	if m.pere != nil {
		mid = m.pere.id
	} else {
		mid = "nil"
	}
	return fmt.Sprintf("modele %s père %s, pos %s sufd %s, %d Genr",
		m.id, mid, m.pos, m.sufd, len(m.lgenR))
}

// habet(d *Des)
// vrai si le modèle m a déjà la désinence d
func (m Modele) habet(d *Des) bool {
	for _, ldes := range m.desm {
		for _, des := range ldes {
			if des.morpho == d.morpho {
				return true
			}
		}
	}
	return false
}

// habetR(gnr Genrad) bool
// vrai si le modèle m a déjà le générateur de radical gnr
func (m Modele) habetR(gnr Genrad) bool {
	for _, genrad := range m.lgenR {
		if genrad.num == gnr.num {
			return true
		}
	}
	return false
}

// renonce(des *Des) bool
// vrai si le modèle m refuse d'hériter de la déninence
// de morpho n° m, parce ce n° figure dans ses abs
func (m Modele) estabs(des *Des) bool {
	for _, i := range m.abs {
		if i == des.morpho {
			return true
		}
	}
	return false
}

func (m *Modele) herite() {
	if m.pere == nil {
		return
	}
	// héritage du pos
	if m.pos == "" {
		m.pos = m.pere.pos
	}
	/*  héritage des générateurs de radicaux */
	for _, genr := range m.pere.lgenR {
		if !m.habetR(genr) {
			m.lgenR = append(m.lgenR, genr)
		}
	}
	// héritage des désinences
	for _, ldesp := range m.pere.desm {
		for _, desp := range ldesp {
			if !m.estabs(desp) && !m.habet(desp) {
				nd := desp.clone()
				nd.modele = m
				m.desm[nd.nr] = append(m.desm[nd.nr], nd)
			}
		}
	}
	/*  héritage des absenst */
}

func (m *Modele) ajsuffd() {
	if m.sufd == "" {
		return
	}
	for _, ld := range m.desm {
		for _, d := range ld {
			d.grq = d.grq + m.sufd
			d.gr = atone(d.grq)
		}
	}
}

func (m *Modele) ajsuff() {
	for _, lsuf := range m.suf {
		ecl := strings.Split(lsuf, ":")
		li := listei(ecl[0])
		suff := ecl[1]
		for _, i := range li {
			for _, ld := range m.desm {
				for _, d := range ld {
					{
						if d.morpho == i {
							nd := d.clone()
							nd.grq = d.grq + suff
							nd.gr = atone(nd.grq)
							m.desm[d.nr] = append(m.desm[d.nr], nd)
						}
					}
				}
			}
		}
	}
}

var modeles = make(map[string]*Modele)
var vardes = make(map[string][]string)

func lismodeles() {
	ll := lignes("data/modeles.la")
	var m *Modele
	for _, l := range ll {
		if l == "" {
			continue
		}
		if strings.HasPrefix(l, "$") {
			l = strings.TrimPrefix(l, "$")
			ecl := strings.Split(l, "=")
			k := ecl[0]
			vardes[k] = strings.Split(ecl[1], ";")
			continue
		}
		ecl := strings.Split(l, ":")
		cle := ecl[0]
		val := strings.TrimPrefix(l, cle+":")
		switch cle {
		case "modele":
			if m != nil {
				m.herite()
				m.ajsuffd()
				m.ajsuff()
				modeles[m.id] = m
			}
			m = new(Modele)
			m.pere = nil
			m.id = val
			m.desm = make(map[int][]*Des)
		case "R":
			num := strtoint(ecl[1])
			if ecl[2] == "K" {
				m.lgenR = append(m.lgenR, Genrad{num, 0, ""})
			} else {
				lp := strings.Split(ecl[2], ",")
				oter := strtoint(lp[0])
				ajout := lp[1]
				if ajout == "0" {
					ajout = ""
				}
				m.lgenR = append(m.lgenR, Genrad{num, oter, ajout})
			}
		case "abs":
			m.abs = listei(ecl[1])
		case "des":
			li := listei(ecl[1])
			// cas des variables
			nr := strtoint(ecl[2])
			ld := ecl[3]
			var dd []string = strings.Split(ld, ";")
			var ddd []string
			for _, sdes := range dd {
				if strings.HasPrefix(sdes, "$") {
					ddv := strings.TrimPrefix(sdes, "$")
					ldv := vardes[ddv]
					for _, dv := range ldv {
						ddd = append(ddd, dv)
					}
				} else if strings.Contains(sdes, "$") {
					fix := strings.Split(sdes, "$")
					prefix := fix[0]
					suffix := fix[1]
					ldv := vardes[suffix]
					for _, dv := range ldv {
						ddd = append(ddd, prefix+dv)
					}
				} else {
					if sdes == "-" {
						sdes = ""
					}
					ddd = append(ddd, sdes)
				}
			}
			maxd := len(ddd)
			var nd *Des
			for ides, ili := range li {
				if ides < maxd {
					sld := ddd[ides]
					ecld := strings.Split(sld, ",")
					for _, cld := range ecld {
						nd = creeDes(cld, m, ili, nr)
						m.desm[nd.nr] = append(m.desm[nd.nr], nd)
					}
				} else if nd != nil {
					nnd := nd.clone()
					nnd.morpho += 1
					m.desm[nd.nr] = append(m.desm[nd.nr], nnd)
				}
			}
		// désinences supplémentaires
		case "des+":
			// TODO les des+ peuvent utiliser les $listes
			if m.pere != nil {
				li := listei(ecl[1])
				for i := 0; i < len(li); i++ {
					ld := m.pere.desm[li[i]]
					for j := 0; j < len(ld); j++ {
						nd := ld[j].clone()
						nd.modele = m
						// XXX attention, cette désinence doit pouvoir être suffixée
						m.desm[nd.nr] = append(m.desm[nd.nr], nd)
					}
				}
			}
		case "pos":
			m.pos = val
		case "suf":
			m.suf = append(m.suf, val)
		case "sufd":
			m.sufd = val
		case "pere":
			m.pere = modeles[val]
		}
	}
	// il faut ajouter le dernier modèle lu
	m.ajsuffd()
	m.ajsuff()
	modeles[m.id] = m
}
