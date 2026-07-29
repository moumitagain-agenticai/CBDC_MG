package fee

// FeeStructure represents the structure of a fee
type FeeStructure string

const (
    StructureFlat      FeeStructure = "FLAT"
    StructurePercentage FeeStructure = "PERCENTAGE"
    StructureTiered    FeeStructure = "TIERED"
    StructureHybrid    FeeStructure = "HYBRID"
)

// String returns the string representation
func (s FeeStructure) String() string {
    return string(s)
}