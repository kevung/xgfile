package xgfile

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type GameDataFormatHdrRecord struct {
	MagicNumber     string
	HeaderVersion   int32
	HeaderSize      int32
	ThumbnailOffset int64
	ThumbnailSize   int32
	GameGUID        string
	GameName        string
	SaveName        string
	LevelName       string
	Comments        string
}

func (hdr *GameDataFormatHdrRecord) FromStream(stream io.Reader) error {
	var unpackedData [8232]byte
	if _, err := io.ReadFull(stream, unpackedData[:]); err != nil {
		return err
	}

	hdr.MagicNumber = string(unpackedData[0:4])
	hdr.HeaderVersion = int32(binary.LittleEndian.Uint32(unpackedData[4:8]))
	if hdr.MagicNumber != "HMGR" || hdr.HeaderVersion != 1 {
		return errors.New("invalid magic number or header version")
	}

	hdr.HeaderSize = int32(binary.LittleEndian.Uint32(unpackedData[8:12]))
	hdr.ThumbnailOffset = int64(binary.LittleEndian.Uint64(unpackedData[12:20]))
	hdr.ThumbnailSize = int32(binary.LittleEndian.Uint32(unpackedData[20:24]))

	guid := unpackedData[24:40]
	hdr.GameGUID = formatGUID(guid)

	hdr.GameName = UTF16IntArrayToStr(unpackUint16Array(unpackedData[40:2064]))
	hdr.SaveName = UTF16IntArrayToStr(unpackUint16Array(unpackedData[2064:4088]))
	hdr.LevelName = UTF16IntArrayToStr(unpackUint16Array(unpackedData[4088:6112]))
	hdr.Comments = UTF16IntArrayToStr(unpackUint16Array(unpackedData[6112:8136]))

	return nil
}

func formatGUID(guid []byte) string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.LittleEndian.Uint32(guid[0:4]),
		binary.LittleEndian.Uint16(guid[4:6]),
		binary.LittleEndian.Uint16(guid[6:8]),
		binary.LittleEndian.Uint16(guid[8:10]),
		guid[10:16])
}

func unpackUint16Array(data []byte) []uint16 {
	var result []uint16
	for i := 0; i < len(data); i += 2 {
		result = append(result, binary.LittleEndian.Uint16(data[i:i+2]))
	}
	return result
}

type TimeSettingRecord struct {
	ClockType    int32
	PerGame      bool
	Time1        int32
	Time2        int32
	Penalty      int32
	TimeLeft1    int32
	TimeLeft2    int32
	PenaltyMoney int32
}

func (tsr *TimeSettingRecord) FromStream(stream io.Reader) error {
	var unpackedData [32]byte
	if _, err := io.ReadFull(stream, unpackedData[:]); err != nil {
		return err
	}

	tsr.ClockType = int32(binary.LittleEndian.Uint32(unpackedData[0:4]))
	tsr.PerGame = unpackedData[4] != 0
	tsr.Time1 = int32(binary.LittleEndian.Uint32(unpackedData[8:12]))
	tsr.Time2 = int32(binary.LittleEndian.Uint32(unpackedData[12:16]))
	tsr.Penalty = int32(binary.LittleEndian.Uint32(unpackedData[16:20]))
	tsr.TimeLeft1 = int32(binary.LittleEndian.Uint32(unpackedData[20:24]))
	tsr.TimeLeft2 = int32(binary.LittleEndian.Uint32(unpackedData[24:28]))
	tsr.PenaltyMoney = int32(binary.LittleEndian.Uint32(unpackedData[28:32]))

	return nil
}

type EvalLevelRecord struct {
	Level    int16
	IsDouble bool
}

func (elr *EvalLevelRecord) FromStream(stream io.Reader) error {
	var unpackedData [4]byte
	if _, err := io.ReadFull(stream, unpackedData[:]); err != nil {
		return err
	}

	elr.Level = int16(binary.LittleEndian.Uint16(unpackedData[0:2]))
	elr.IsDouble = unpackedData[2] != 0

	return nil
}

type EngineStructBestMoveRecord struct {
	Pos       [26]int8
	Dice      [2]int8
	Level     int32
	Score     [2]int32
	Cube      int32
	CubePos   int32
	Crawford  int32
	Jacoby    int32
	NMoves    int32
	PosPlayed [32][26]int8
	Moves     [32][8]int8
	EvalLevel [32]EvalLevelRecord
	Eval      [32][7]float32
	Unused    int8
	Met       int8
	Choice0   int8
	Choice3   int8
}

func (esbmr *EngineStructBestMoveRecord) FromStream(stream io.Reader) error {
	var unpackedData [68]byte
	if _, err := io.ReadFull(stream, unpackedData[:]); err != nil {
		return err
	}

	copy(esbmr.Pos[:], unpackedData[0:26])
	copy(esbmr.Dice[:], unpackedData[26:28])
	esbmr.Level = int32(binary.LittleEndian.Uint32(unpackedData[28:32]))
	esbmr.Score[0] = int32(binary.LittleEndian.Uint32(unpackedData[32:36]))
	esbmr.Score[1] = int32(binary.LittleEndian.Uint32(unpackedData[36:40]))
	esbmr.Cube = int32(binary.LittleEndian.Uint32(unpackedData[40:44]))
	esbmr.CubePos = int32(binary.LittleEndian.Uint32(unpackedData[44:48]))
	esbmr.Crawford = int32(binary.LittleEndian.Uint32(unpackedData[48:52]))
	esbmr.Jacoby = int32(binary.LittleEndian.Uint32(unpackedData[52:56]))
	esbmr.NMoves = int32(binary.LittleEndian.Uint32(unpackedData[56:60]))

	for i := 0; i < 32; i++ {
		if _, err := io.ReadFull(stream, esbmr.PosPlayed[i][:]); err != nil {
			return err
		}
	}

	for i := 0; i < 32; i++ {
		if _, err := io.ReadFull(stream, esbmr.Moves[i][:]); err != nil {
			return err
		}
	}

	for i := 0; i < 32; i++ {
		if err := esbmr.EvalLevel[i].FromStream(stream); err != nil {
			return err
		}
	}

	for i := 0; i < 32; i++ {
		var evalData [28]byte
		if _, err := io.ReadFull(stream, evalData[:]); err != nil {
			return err
		}
		for j := 0; j < 7; j++ {
			esbmr.Eval[i][j] = float32(binary.LittleEndian.Uint32(evalData[j*4 : (j+1)*4]))
		}
	}

	var finalData [4]byte
	if _, err := io.ReadFull(stream, finalData[:]); err != nil {
		return err
	}
	esbmr.Unused = int8(finalData[0])
	esbmr.Met = int8(finalData[1])
	esbmr.Choice0 = int8(finalData[2])
	esbmr.Choice3 = int8(finalData[3])

	return nil
}

type EngineStructDoubleAction struct {
	Pos           [26]int8
	Level         int32
	Score         [2]int32
	Cube          int32
	CubePos       int32
	Jacoby        int32
	Crawford      int32
	Met           int32
	FlagDouble    int32
	IsBeaver      int32
	Eval          [7]float32
	EquB          float32
	EquDouble     float32
	EquDrop       float32
	LevelRequest  int32
	DoubleChoice3 int32
	EvalDouble    [7]float32
}

func (esdar *EngineStructDoubleAction) FromStream(stream io.Reader) error {
	var unpackedData [132]byte
	if _, err := io.ReadFull(stream, unpackedData[:]); err != nil {
		return err
	}

	copy(esdar.Pos[:], unpackedData[0:26])
	esdar.Level = int32(binary.LittleEndian.Uint32(unpackedData[26:30]))
	esdar.Score[0] = int32(binary.LittleEndian.Uint32(unpackedData[30:34]))
	esdar.Score[1] = int32(binary.LittleEndian.Uint32(unpackedData[34:38]))
	esdar.Cube = int32(binary.LittleEndian.Uint32(unpackedData[38:42]))
	esdar.CubePos = int32(binary.LittleEndian.Uint32(unpackedData[42:46]))
	esdar.Jacoby = int32(binary.LittleEndian.Uint32(unpackedData[46:50]))
	esdar.Crawford = int32(binary.LittleEndian.Uint32(unpackedData[50:54]))
	esdar.Met = int32(binary.LittleEndian.Uint32(unpackedData[54:58]))
	esdar.FlagDouble = int32(binary.LittleEndian.Uint32(unpackedData[58:62]))
	esdar.IsBeaver = int32(binary.LittleEndian.Uint32(unpackedData[62:66]))
	for i := 0; i < 7; i++ {
		esdar.Eval[i] = float32(binary.LittleEndian.Uint32(unpackedData[66+i*4 : 70+i*4]))
	}
	esdar.EquB = float32(binary.LittleEndian.Uint32(unpackedData[94:98]))
	esdar.EquDouble = float32(binary.LittleEndian.Uint32(unpackedData[98:102]))
	esdar.EquDrop = float32(binary.LittleEndian.Uint32(unpackedData[102:106]))
	esdar.LevelRequest = int32(binary.LittleEndian.Uint32(unpackedData[106:110]))
	esdar.DoubleChoice3 = int32(binary.LittleEndian.Uint32(unpackedData[110:114]))
	for i := 0; i < 7; i++ {
		esdar.EvalDouble[i] = float32(binary.LittleEndian.Uint32(unpackedData[114+i*4 : 118+i*4]))
	}

	return nil
}

type HeaderMatchEntry struct {
	Name                 string
	EntryType            int32
	SPlayer1             string
	SPlayer2             string
	MatchLength          int32
	Variation            int32
	Crawford             bool
	Jacoby               bool
	Beaver               bool
	AutoDouble           bool
	Elo1                 float64
	Elo2                 float64
	Exp1                 int32
	Exp2                 int32
	Date                 string
	SEvent               string
	GameId               int32
	CompLevel1           int32
	CompLevel2           int32
	CountForElo          bool
	AddtoProfile1        bool
	AddtoProfile2        bool
	SLocation            string
	GameMode             int32
	Imported             bool
	SRound               string
	Invert               int32
	Version              int32
	Magic                int32
	MoneyInitG           int32
	MoneyInitScore       [2]int32
	Entered              bool
	Counted              bool
	UnratedImp           bool
	CommentHeaderMatch   int32
	CommentFooterMatch   int32
	IsMoneyMatch         bool
	WinMoney             float64
	LoseMoney            float64
	Currency             int32
	FeeMoney             float64
	TableStake           int32
	SiteId               int32
	CubeLimit            int32
	AutoDoubleMax        int32
	Transcribed          bool
	Event                string
	Player1              string
	Player2              string
	Location             string
	Round                string
	TimeSetting          TimeSettingRecord
	TotTimeDelayMove     int32
	TotTimeDelayCube     int32
	TotTimeDelayMoveDone int32
	TotTimeDelayCubeDone int32
	Transcriber          string
}

func (hme *HeaderMatchEntry) FromStream(stream io.Reader) error {
	// Implement the method based on the Python version
	return nil
}

type FooterGameEntry struct {
	Name           string
	EntryType      int32
	Score1g        int32
	Score2g        int32
	CrawfordApplyg bool
	Winner         int32
	PointsWon      int32
	Termination    int32
	ErrResign      float64
	ErrTakeResign  float64
	Eval           [7]float32
	EvalLevel      int32
}

func (fge *FooterGameEntry) FromStream(stream io.Reader) error {
	// Implement the method based on the Python version
	return nil
}

type MissingEntry struct {
	Name           string
	EntryType      int32
	MissingErrLuck float64
	MissingWinner  int32
	MissingPoints  int32
}

func (me *MissingEntry) FromStream(stream io.Reader) error {
	// Implement the method based on the Python version
	return nil
}

type FooterMatchEntry struct {
	Name      string
	EntryType int32
	Score1m   int32
	Score2m   int32
	WinnerM   int32
	Elo1m     float64
	Elo2m     float64
	Exp1m     int32
	Exp2m     int32
	Datem     string
}

func (fme *FooterMatchEntry) FromStream(stream io.Reader) error {
	// Implement the method based on the Python version
	return nil
}

type HeaderGameEntry struct {
	Name                string
	EntryType           int32
	Score1              int32
	Score2              int32
	CrawfordApply       bool
	PosInit             [26]int8
	GameNumber          int32
	InProgress          bool
	CommentHeaderGame   int32
	CommentFooterGame   int32
	NumberOfAutoDoubles int32
}

func (hge *HeaderGameEntry) FromStream(stream io.Reader) error {
	// Implement the method based on the Python version
	return nil
}

type CubeEntry struct {
	Name                   string
	EntryType              int32
	ActiveP                int32
	Double                 int32
	Take                   int32
	BeaverR                int32
	RaccoonR               int32
	CubeB                  int32
	Position               [26]int8
	Doubled                EngineStructDoubleAction
	ErrCube                float64
	DiceRolled             string
	ErrTake                float64
	RolloutIndexD          int32
	CompChoiceD            int32
	AnalyzeC               int32
	ErrBeaver              float64
	ErrRaccoon             float64
	AnalyzeCR              int32
	IsValid                int32
	TutorCube              int32
	TutorTake              int32
	ErrTutorCube           float64
	ErrTutorTake           float64
	FlaggedDouble          bool
	CommentCube            int32
	EditedCube             bool
	TimeDelayCube          bool
	TimeDelayCubeDone      bool
	NumberOfAutoDoubleCube int32
	TimeBot                int32
	TimeTop                int32
}

func (ce *CubeEntry) FromStream(stream io.Reader) error {
	// Implement the method based on the Python version
	return nil
}

type MoveEntry struct {
	Name                   string
	EntryType              int32
	PositionI              [26]int8
	PositionEnd            [26]int8
	ActiveP                int32
	Moves                  [8]int32
	Dice                   [2]int32
	CubeA                  int32
	ErrorM                 int32
	NMoveEval              int32
	DataMoves              EngineStructBestMoveRecord
	Played                 bool
	ErrMove                float64
	ErrLuck                float64
	CompChoice             int32
	InitEq                 float64
	RolloutIndexM          [32]int32
	AnalyzeM               int32
	AnalyzeL               int32
	InvalidM               int32
	PositionTutor          [26]int8
	Tutor                  int32
	ErrTutorMove           float64
	Flagged                bool
	CommentMove            int32
	EditedMove             bool
	TimeDelayMove          int32
	TimeDelayMoveDone      int32
	NumberOfAutoDoubleMove int32
}

func (me *MoveEntry) FromStream(stream io.Reader) error {
	// Implement the method based on the Python version
	return nil
}

type UnimplementedEntry struct {
	Name string
}

func (ue *UnimplementedEntry) FromStream(stream io.Reader) error {
	// Implement the method based on the Python version
	return nil
}

type GameFileRecord struct {
	Name      string
	EntryType int32
	Record    interface{}
	Version   int32
}

func (gfr *GameFileRecord) FromStream(stream io.Reader) error {
	// Implement the method based on the Python version
	return nil
}

type RolloutContextEntry struct {
	Name                 string
	EntryType            int32
	Truncated            bool
	ErrorLimited         bool
	Truncate             int32
	MinRoll              int32
	ErrorLimit           float64
	MaxRoll              int32
	Level1               int32
	Level2               int32
	LevelCut             int32
	Variance             bool
	Cubeless             bool
	Time                 bool
	Level1C              int32
	Level2C              int32
	TimeLimit            int32
	TruncateBO           int32
	RandomSeed           int32
	RandomSeedI          int32
	RollBoth             bool
	SearchInterval       float64
	Met                  int32
	FirstRoll            bool
	DoDouble             bool
	Extent               bool
	Rolled               int32
	DoubleFirst          bool
	Sum1                 [37]float64
	SumSquare1           [37]float64
	Sum2                 [37]float64
	SumSquare2           [37]float64
	Stdev1               [37]float64
	Stdev2               [37]float64
	RolledD              [37]int32
	Error1               float64
	Error2               float64
	Result1              [7]float32
	Result2              [7]float32
	Mwc1                 float64
	Mwc2                 float64
	PrevLevel            int32
	PrevEval             [7]float32
	PrevND               int32
	PrevD                int32
	Duration             int32
	LevelTrunc           int32
	Rolled2              int32
	MultipleMin          int32
	MultipleStopAll      bool
	MultipleStopOne      bool
	MultipleStopAllValue float64
	MultipleStopOneValue float64
	AsTake               bool
	Rotation             int32
	UserInterrupted      bool
	VerMaj               int32
	VerMin               int32
}

func (rce *RolloutContextEntry) FromStream(stream io.Reader) error {
	// Implement the method based on the Python version
	return nil
}

type RolloutFileRecord struct {
	Name      string
	EntryType int32
	Record    RolloutContextEntry
	Version   int32
}

func (rfr *RolloutFileRecord) FromStream(stream io.Reader) error {
	// Implement the method based on the Python version
	return nil
}

// Define other structures and methods similarly...
