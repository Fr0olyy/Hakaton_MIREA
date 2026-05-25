from typing import Dict, Any

def evaluate_dataset_policy(tabular_score: float, cv_score: float) -> Dict[str, Any]:
    """
    Принимает финальные скоры от табличного и CV адаптеров.
    На основе правил (Policy) принимает бизнес-решение о дальнейшей судьбе датасета.
    """
    
    # ШАГ 1: Подготовка
    # Вычисляем средний скор по двум модальностям
    # Сложи tabular_score и cv_score, затем раздели на 2
    overall_score = (tabular_score+cv_score)/2
    
    # ШАГ 2: Логика принятия решений (Policy Engine)
    # Заведи переменную decision со строкой "unknown" по умолчанию
    decision = "unknown"
    
    # Напиши ветвление if-elif-else:
    # 1. Если overall_score больше или равен 85.0 -> decision = "accept" (одобрено)
    # 2. Если overall_score от 60.0 до 84.9 -> decision = "augment" (отправить на аугментацию/синтетику)
    # 3. Во всех остальных случаях -> decision = "reject" (забраковать датасет)
    if overall_score>=85.0:
        decision = "accept"
    elif 60.0<=overall_score<=84.9:
        decision = "augment"
    else:
        decision = "reject"


    result = {
        "overall_quality_score": overall_score,
        "final_decision": decision
    }
    
    return result