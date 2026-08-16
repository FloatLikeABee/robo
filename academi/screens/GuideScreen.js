import React, { useState, useMemo } from 'react';
import {
  View, Text, StyleSheet, FlatList, TouchableOpacity,
  Dimensions, ScrollView
} from 'react-native';
import Button from '../components/Button';

const { width } = Dimensions.get('window');

const mockGuides = [
  {
    id: 1,
    title: "Write a Research Paper",
    description: "Step-by-step guide to crafting an academic research paper",
    steps: 7,
    progress: 0,
    category: "Writing",
    icon: "📝",
    color: "#5B8CFF",
    stepItems: [
      "Choose a research topic",
      "Conduct preliminary research",
      "Create an outline",
      "Write a thesis statement",
      "Draft the introduction",
      "Write body paragraphs",
      "Write conclusion and edit",
    ],
  },
  {
    id: 2,
    title: "Learn Calculus in 7 Days",
    description: "Master calculus concepts with daily practice problems",
    steps: 7,
    progress: 0,
    category: "Mathematics",
    icon: "📊",
    color: "#9B6DFF",
    stepItems: [
      "Day 1: Limits and continuity",
      "Day 2: Derivative definition",
      "Day 3: Differentiation rules",
      "Day 4: Applications of derivatives",
      "Day 5: Introduction to integrals",
      "Day 6: Integration techniques",
      "Day 7: Review problem set",
    ],
  },
  {
    id: 3,
    title: "Prepare for SAT Exam",
    description: "Comprehensive SAT preparation strategy and practice",
    steps: 10,
    progress: 0,
    category: "Test Prep",
    icon: "📚",
    color: "#00D4FF",
    stepItems: [
      "Diagnostic practice test",
      "Reading: evidence questions",
      "Writing: grammar rules",
      "Math: algebra fundamentals",
      "Math: problem solving & data",
      "Full practice test 1",
      "Review missed questions",
      "Timed section drills",
      "Essay structure (if applicable)",
      "Final confidence review",
    ],
  },
];

function buildInitialSteps(guide) {
  const titles = guide.stepItems && guide.stepItems.length ? guide.stepItems : Array.from({ length: guide.steps || 0 }, (_, i) => `Step ${i + 1}`);
  return titles.map((title, i) => ({
    id: i + 1,
    title,
    completed: false,
  }));
}

export default function GuideScreen() {
  const [activeGuide, setActiveGuide] = useState(null);
  /** @type {{ [guideId: string]: { id: number, title: string, completed: boolean }[] }} */
  const [stepsByGuideId, setStepsByGuideId] = useState({});

  const startGuide = (guide) => {
    setActiveGuide(guide);
    setStepsByGuideId((prev) => {
      if (prev[guide.id]) return prev;
      return { ...prev, [guide.id]: buildInitialSteps(guide) };
    });
  };

  const guideSteps = activeGuide ? stepsByGuideId[activeGuide.id] || [] : [];

  const toggleStep = (guideId, stepId) => {
    setStepsByGuideId((prev) => ({
      ...prev,
      [guideId]: (prev[guideId] || []).map((step) =>
        step.id === stepId ? { ...step, completed: !step.completed } : step,
      ),
    }));
  };

  const completedCount = useMemo(
    () => guideSteps.filter((step) => step.completed).length,
    [guideSteps],
  );

  const totalSteps = guideSteps.length || 1;

  const currentProgress = useMemo(
    () => Math.round((completedCount / totalSteps) * 100),
    [completedCount, totalSteps],
  );

  const listProgress = (guide) => {
    const list = stepsByGuideId[guide.id];
    const total = list?.length || guide.steps || 1;
    const done = list?.filter((s) => s.completed).length || 0;
    const pct = total ? Math.round((done / total) * 100) : 0;
    return { done, total, pct };
  };

  if (activeGuide) {
    return (
      <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
        <View style={styles.guideHeader}>
          <TouchableOpacity onPress={() => setActiveGuide(null)} style={styles.backButton}>
            <Text style={styles.backButtonText}>← Back</Text>
          </TouchableOpacity>
          <View style={styles.guideTitleRow}>
            <Text style={styles.guideTitleIcon}>{activeGuide.icon}</Text>
            <Text style={styles.guideTitle}>{activeGuide.title}</Text>
          </View>
          <Text style={styles.guideDescription}>{activeGuide.description}</Text>
        </View>

          <View style={styles.progressSection}>
          <View style={styles.progressRingOuter}>
            <View style={[styles.progressRing, { borderColor: activeGuide.color }]}>
              <Text style={[styles.progressText, { color: activeGuide.color }]}>
                {completedCount}/{totalSteps}
              </Text>
            </View>
          </View>
          <Text style={styles.progressLabel}>
            {completedCount} of {totalSteps} steps ({currentProgress}%)
          </Text>
        </View>

        <View style={styles.timelineContainer}>
          {guideSteps.map((step, index) => (
            <TouchableOpacity
              key={step.id}
              style={[styles.stepItem, step.completed && styles.stepCompleted]}
              onPress={() => toggleStep(activeGuide.id, step.id)}
              activeOpacity={0.7}
            >
              <View style={styles.timelineLine}>
                {index < guideSteps.length - 1 && (
                  <View style={[styles.timelineConnector, step.completed && styles.timelineConnectorActive]} />
                )}
              </View>
              <View style={[styles.stepNumberCircle, step.completed && styles.stepNumberCircleCompleted]}>
                <Text style={[styles.stepNumber, step.completed && styles.stepNumberCompleted]}>
                  {step.completed ? "✓" : step.id}
                </Text>
              </View>
              <View style={styles.stepContent}>
                <Text style={[styles.stepTitle, step.completed && styles.stepTitleCompleted]}>
                  {step.title}
                </Text>
              </View>
            </TouchableOpacity>
          ))}
        </View>

        <View style={styles.actionsContainer}>
          <Button variant="ghost" label="Back to Guides" onPress={() => setActiveGuide(null)} />
          <Button variant="primary" label="Get AI Feedback" onPress={() => {}} />
        </View>
      </ScrollView>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Guide</Text>
        <Text style={styles.headerSubtitle}>Learn and accomplish goals</Text>
      </View>

      <FlatList
        data={mockGuides}
        keyExtractor={item => item.id.toString()}
        contentContainerStyle={styles.guidesContentContainer}
        renderItem={({ item }) => {
          const { done, total, pct } = listProgress(item);
          return (
            <TouchableOpacity
              style={styles.guideCard}
              onPress={() => startGuide(item)}
              activeOpacity={0.8}
            >
              <View style={styles.guideCardContent}>
                <View style={styles.guideCardHeader}>
                  <Text style={styles.guideCardIcon}>{item.icon}</Text>
                  <View style={styles.guideCardInfo}>
                    <Text style={styles.guideCardTitle}>{item.title}</Text>
                    <Text style={styles.guideCardCategory}>{item.category}</Text>
                  </View>
                </View>
                <Text style={styles.guideCardDescription}>{item.description}</Text>
                <View style={styles.guideCardFooter}>
                  <Text style={styles.guideCardSteps}>
                    {done}/{total}
                  </Text>
                  <View style={styles.progressBarContainer}>
                    <View style={[styles.progressBar, { width: `${pct}%`, backgroundColor: item.color }]} />
                  </View>
                  <Text style={styles.progressPercentage}>{pct}%</Text>
                </View>
              </View>
            </TouchableOpacity>
          );
        }}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0A0F1C',
  },
  header: {
    padding: 24,
    paddingTop: 48,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255,255,255,0.08)',
  },
  headerTitle: {
    color: '#FFFFFF',
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 4,
  },
  headerSubtitle: {
    color: '#A8B2D1',
    fontSize: 14,
  },
  guidesContentContainer: {
    padding: 16,
    paddingBottom: 80,
  },
  guideCard: {
    backgroundColor: 'rgba(255,255,255,0.02)',
    borderRadius: 16,
    overflow: 'hidden',
    marginBottom: 16,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.08)',
  },
  guideCardContent: {
    padding: 16,
  },
  guideCardHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  guideCardIcon: {
    fontSize: 28,
    marginRight: 12,
  },
  guideCardInfo: {
    flex: 1,
  },
  guideCardTitle: {
    color: '#FFFFFF',
    fontSize: 18,
    fontWeight: '600',
    marginBottom: 4,
  },
  guideCardCategory: {
    color: '#A8B2D1',
    fontSize: 12,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  guideCardDescription: {
    color: '#A8B2D1',
    fontSize: 14,
    marginBottom: 16,
    lineHeight: 20,
  },
  guideCardFooter: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  guideCardSteps: {
    color: '#A8B2D1',
    fontSize: 12,
    marginRight: 12,
  },
  progressBarContainer: {
    flex: 1,
    height: 6,
    backgroundColor: 'rgba(255,255,255,0.1)',
    borderRadius: 3,
    overflow: 'hidden',
  },
  progressBar: {
    height: '100%',
    borderRadius: 3,
  },
  progressPercentage: {
    color: '#FFFFFF',
    fontSize: 12,
    fontWeight: '600',
    marginLeft: 12,
    width: 36,
    textAlign: 'right',
  },
  guideHeader: {
    padding: 24,
    paddingTop: 48,
  },
  backButton: {
    marginBottom: 16,
  },
  backButtonText: {
    color: '#5B8CFF',
    fontSize: 16,
    fontWeight: '500',
  },
  guideTitleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  guideTitleIcon: {
    fontSize: 32,
    marginRight: 12,
  },
  guideTitle: {
    color: '#FFFFFF',
    fontSize: 22,
    fontWeight: '600',
    flex: 1,
  },
  guideDescription: {
    color: '#A8B2D1',
    fontSize: 16,
    marginBottom: 24,
    lineHeight: 22,
  },
  progressSection: {
    alignItems: 'center',
    padding: 24,
    marginBottom: 16,
  },
  progressRingOuter: {
    marginBottom: 12,
  },
  progressRing: {
    width: 100,
    height: 100,
    borderRadius: 50,
    borderWidth: 8,
    borderColor: '#5B8CFF',
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: 'rgba(255,255,255,0.02)',
  },
  progressText: {
    fontSize: 28,
    fontWeight: 'bold',
  },
  progressLabel: {
    color: '#A8B2D1',
    fontSize: 14,
  },
  timelineContainer: {
    padding: 16,
    paddingHorizontal: 24,
  },
  stepItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 16,
  },
  timelineLine: {
    alignItems: 'center',
    width: 32,
    marginRight: 8,
  },
  timelineConnector: {
    width: 2,
    height: 40,
    backgroundColor: 'rgba(255,255,255,0.1)',
  },
  timelineConnectorActive: {
    backgroundColor: '#5B8CFF',
  },
  stepNumberCircle: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: 'rgba(255,255,255,0.08)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  stepNumberCircleCompleted: {
    backgroundColor: '#5B8CFF',
  },
  stepNumber: {
    color: '#A8B2D1',
    fontSize: 14,
    fontWeight: '600',
  },
  stepNumberCompleted: {
    color: '#FFFFFF',
  },
  stepContent: {
    flex: 1,
    paddingLeft: 12,
  },
  stepTitle: {
    color: '#FFFFFF',
    fontSize: 16,
  },
  stepTitleCompleted: {
    color: '#A8B2D1',
    textDecorationLine: 'line-through',
  },
  stepCompleted: {
    opacity: 0.7,
  },
  actionsContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    padding: 24,
  },
});
