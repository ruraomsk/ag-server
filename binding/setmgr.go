package binding

type MGR struct {
	Phase int `json:"phase"` //Номер фазы за которой разрена вставка МГР
	TLen  int `json:"tlen"`  //Минимальная длительность фазы после кторой можно вставлять МГР
	TMGR  int `json:"tmgr"`
}
