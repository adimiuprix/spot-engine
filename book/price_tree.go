package book

import (
	"github.com/google/btree"
	"github.com/shopspring/decimal"
)

type PriceTree struct {
	tree *btree.BTreeG[PriceItem]

	descending bool
}

func NewPriceTree(descending bool) *PriceTree {

	return &PriceTree{

		tree: btree.NewG(
			32,
			func(a, b PriceItem) bool {

				if descending {

					return a.Price.GreaterThan(
						b.Price,
					)

				}

				return a.Price.LessThan(
					b.Price,
				)
			},
		),

		descending: descending,
	}
}

func (p *PriceTree) Add(
	level *PriceLevel,

) {

	item := PriceItem{

		Price: level.Price,

		Level: level,
	}

	p.tree.ReplaceOrInsert(item)

}

func (p *PriceTree) Get(
	price decimal.Decimal,

) *PriceLevel {

	var result *PriceLevel

	p.tree.Ascend(
		func(item PriceItem) bool {

			if item.Price.Equal(price) {

				result = item.Level

				return false
			}

			return true

		},
	)

	return result
}
